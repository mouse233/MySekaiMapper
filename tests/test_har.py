"""har 模块纯函数与 Reqable 上报端点测试。"""
import base64
import gzip
import json

import msgpack
import pytest
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from fastapi.testclient import TestClient

from app import config, har
from app import server

KEY = b"0123456789abcdef"
IV = b"fedcba9876543210"

REPORT_URL = "/reqable/report"


# ---------- 工具:构造合法加密存档与 HAR ----------

def _encrypt(plain: bytes) -> bytes:
    """与游戏相同的 AES-128-CBC + 尾部填充加密(参考 test_crypto)。"""
    pad = 16 - len(plain) % 16
    data = plain + bytes([pad]) * pad
    cipher = Cipher(algorithms.AES(KEY), modes.CBC(IV))
    enc = cipher.encryptor()
    return enc.update(data) + enc.finalize()


def _make_archive() -> bytes:
    plain = msgpack.packb({"updatedResources": {"userMysekaiHarvestMaps": []}})
    return _encrypt(plain)


def _make_har(url: str, body: bytes, base64_body: bool = True) -> dict:
    content = {"size": len(body), "mimeType": "application/octet-stream"}
    if base64_body:
        content["encoding"] = "base64"
        content["text"] = base64.b64encode(body).decode("ascii")
    else:
        content["text"] = body.decode("utf-8", errors="replace")
    return {"log": {"version": "1.2", "entries": [
        {
            "request": {"method": "GET", "url": url, "headers": [], "postData": {}},
            "response": {"status": 200, "headers": [], "content": content},
        }
    ]}}


# ---------- har 模块纯函数 ----------

def test_decompress_identity():
    assert har.decompress_body(b"hello") == b"hello"
    assert har.decompress_body(b"hello", "identity") == b"hello"


def test_decompress_gzip():
    raw = gzip.compress(b"hello gzip")
    assert har.decompress_body(raw, "gzip") == b"hello gzip"


@pytest.mark.parametrize("encoding", ["br", "zstd"])
def test_decompress_br_zstd(encoding):
    import brotli
    import zstandard

    if encoding == "br":
        raw = brotli.compress(b"hello brotli")
        plain = b"hello brotli"
    else:
        raw = zstandard.ZstdCompressor().compress(b"hello zstd")
        plain = b"hello zstd"
    assert har.decompress_body(raw, encoding) == plain


def test_decompress_zstd_without_content_size():
    """回归:Reqable 流式压缩的 zstd 帧头不携带内容大小,
    必须用 decompressobj() 而非 ZstdDecompressor().decompress()。"""
    import zstandard

    plain = b"hello zstd no-size" * 1000
    raw = zstandard.ZstdCompressor(write_content_size=False).compress(plain)
    assert har.decompress_body(raw, "zstd") == plain


def test_parse_har_falls_back_to_uncompressed_json():
    """回归:带 Content-Encoding 但实际未压缩的请求体也应能解析。"""
    har_data = {"log": {"entries": []}}
    raw = json.dumps(har_data).encode("utf-8")
    assert har.parse_har(raw, "gzip") == har_data


def test_decompress_unknown_raises():
    with pytest.raises(ValueError, match="Unsupported"):
        har.decompress_body(b"x", "deflate")


def test_parse_har_compressed():
    har_data = {"log": {"entries": []}}
    raw = gzip.compress(json.dumps(har_data).encode("utf-8"))
    assert har.parse_har(raw, "gzip") == har_data


def test_har_entries_empty_and_normal():
    assert har.har_entries({"log": {"entries": []}}) == []
    assert har.har_entries({}) == []
    assert har.har_entries({"log": {"entries": [1, 2]}}) == [1, 2]


def test_content_to_bytes_base64():
    obj = {"encoding": "base64", "text": base64.b64encode(b"\x00\x01\xff").decode()}
    assert har.content_to_bytes(obj) == b"\x00\x01\xff"


def test_content_to_bytes_plain_text():
    assert har.content_to_bytes({"text": "hello"}) == b"hello"


def test_content_to_bytes_missing_or_empty():
    assert har.content_to_bytes(None) is None
    assert har.content_to_bytes({}) is None
    assert har.content_to_bytes({"text": None}) is None


def test_entry_candidate_bodies_prefers_response():
    entry = {
        "request": {"postData": {"text": "req-body"}},
        "response": {"content": {"encoding": "base64", "text": base64.b64encode(b"resp").decode()}},
    }
    bodies = list(har.entry_candidate_bodies(entry))
    # 响应体优先,请求体兜底,两者都会被产出
    assert bodies == [b"resp", b"req-body"]

    entry_no_response = {"request": {"postData": {"text": "req-body"}}}
    assert list(har.entry_candidate_bodies(entry_no_response)) == [b"req-body"]


# ---------- 上报端点 ----------

@pytest.fixture
def client(monkeypatch, tmp_path):
    monkeypatch.setenv("AES_KEY", KEY.decode("ascii"))
    monkeypatch.setenv("AES_IV", IV.decode("ascii"))
    monkeypatch.setattr(config, "RAW_DIR", tmp_path)

    async def _noop_run(bin_path, task_id, user_id):
        pass

    monkeypatch.setattr(server, "_run_generate_and_notify", _noop_run)
    return TestClient(server.app)


def _post_har(client, har_data, content_encoding=None, extra_headers=None, raw=None):
    body = raw if raw is not None else json.dumps(har_data).encode("utf-8")
    if content_encoding == "gzip":
        body = gzip.compress(body)
    elif content_encoding == "br":
        import brotli
        body = brotli.compress(body)
    elif content_encoding == "zstd":
        import zstandard
        body = zstandard.ZstdCompressor().compress(body)
    headers = {"Content-Type": "application/json"}
    if content_encoding:
        headers["Content-Encoding"] = content_encoding
    headers.update(extra_headers or {})
    return client.post(REPORT_URL, content=body, headers=headers)


def test_report_disabled_returns_404(client, monkeypatch):
    monkeypatch.setattr(config, "REPORT_ENABLED", False)
    r = client.post(REPORT_URL, content=b"{}")
    assert r.status_code == 404


def test_report_accepts_gzip_har(client, tmp_path):
    archive = _make_archive()
    har_data = _make_har("https://api.example.com/user/1234567890/mysekai", archive)
    r = _post_har(
        client, har_data, content_encoding="gzip",
        extra_headers={
            "x-reqable-platform": "android",
            "x-reqable-reporter-host": "api.example.com",
            "x-reqable-reporter-rule": "https://api.example.com/user/*/mysekai*",
        },
    )
    assert r.status_code == 200
    assert r.text == "ok"
    files = list(tmp_path.glob("mysekai_*.bin"))
    assert len(files) == 1
    assert files[0].name.startswith("mysekai_1234567890_")
    assert files[0].read_bytes() == archive


@pytest.mark.parametrize("encoding", ["br", "zstd"])
def test_report_accepts_br_zstd(client, tmp_path, encoding):
    archive = _make_archive()
    har_data = _make_har("https://api.example.com/user/42/mysekai", archive)
    r = _post_har(client, har_data, content_encoding=encoding)
    assert r.status_code == 200
    files = list(tmp_path.glob("mysekai_42_*.bin"))
    assert len(files) == 1


def test_report_skips_non_archive_entries(client, tmp_path):
    archive = _make_archive()
    # 第一条是普通 API 会话(不可解密),第二条才是存档
    har_data = {
        "log": {"entries": [
            {
                "request": {"method": "GET", "url": "https://api.example.com/ping"},
                "response": {"status": 200, "content": {"mimeType": "application/json", "text": "{}"}},
            },
            {
                "request": {"method": "GET", "url": "https://api.example.com/user/777/mysekai"},
                "response": {"status": 200, "content": {
                    "encoding": "base64",
                    "text": base64.b64encode(archive).decode("ascii"),
                }},
            },
        ]}
    }
    r = _post_har(client, har_data)
    assert r.status_code == 200
    files = list(tmp_path.glob("mysekai_*.bin"))
    assert len(files) == 1
    assert files[0].name.startswith("mysekai_777_")


def test_report_no_valid_entry_returns_ok_no_file(client, tmp_path):
    har_data = {
        "log": {"entries": [
            {
                "request": {"method": "GET", "url": "https://api.example.com/ping"},
                "response": {"status": 200, "content": {"mimeType": "application/json", "text": "{}"}},
            }
        ]}
    }
    r = _post_har(client, har_data)
    assert r.status_code == 200
    assert list(tmp_path.glob("*.bin")) == []


def test_report_token_required(client, tmp_path, monkeypatch):
    monkeypatch.setattr(config, "REPORT_TOKEN", "s3cret")
    archive = _make_archive()
    har_data = _make_har("https://api.example.com/user/1/mysekai", archive)

    r = _post_har(client, har_data)
    assert r.status_code == 401

    r = _post_har(client, har_data, extra_headers={"X-Report-Token": "wrong"})
    assert r.status_code == 401

    r = _post_har(client, har_data, extra_headers={"X-Report-Token": "s3cret"})
    assert r.status_code == 200
    assert len(list(tmp_path.glob("*.bin"))) == 1


def test_report_invalid_body_returns_400(client):
    r = _post_har(client, None, raw=b"not json at all")
    assert r.status_code == 400


def test_report_body_too_large_returns_413(client, monkeypatch):
    monkeypatch.setattr(config, "REPORT_MAX_SIZE", 100)
    r = client.post(REPORT_URL, content=b"x" * 101)
    assert r.status_code == 413
