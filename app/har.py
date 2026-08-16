"""Reqable 上报服务器 HAR 解析(纯函数,便于单测)。

Reqable 的「上报服务器」功能把每个已完成的 HTTP 会话按 HAR 格式(JSON)
POST 到用户自建服务端,可选 gzip / brotli / zstd 压缩(通过 Content-Encoding 标识)。
本模块负责:解压请求体 -> 解析 HAR JSON -> 从会话中提取原始响应/请求字节。
"""
import base64
import gzip
import json
from typing import Iterator, Optional

# HAR 中二进制内容(响应 content / 请求 postData)的 base64 标识
ENCODING_BASE64 = "base64"


def decompress_body(raw: bytes, content_encoding: Optional[str] = None) -> bytes:
    """按 Content-Encoding 解压请求体;无编码或 identity 时原样返回。"""
    enc = (content_encoding or "").strip().lower()
    if not enc or enc in ("identity",):
        return raw
    if enc in ("gzip", "x-gzip"):
        return gzip.decompress(raw)
    if enc == "br":
        import brotli

        return brotli.decompress(raw)
    if enc in ("zstd", "zstandard"):
        import zstandard

        # 注意:Reqable(流式压缩)产出的 zstd 帧头通常不携带内容大小,
        # ZstdDecompressor().decompress() 会报
        # "could not determine content size in frame header",
        # 因此必须用不依赖帧头 content size 的 decompressobj()。
        return zstandard.ZstdDecompressor().decompressobj().decompress(raw)
    raise ValueError(f"Unsupported Content-Encoding: {content_encoding!r}")


def parse_har(raw: bytes, content_encoding: Optional[str] = None) -> dict:
    """解压并解析 HAR JSON。

    解压失败时兜底按未压缩 JSON 直接解析(Reqable 偶发会带编码头却发未压缩体,
    失败不重试,尽量宽容)。
    """
    try:
        body = decompress_body(raw, content_encoding)
    except Exception:
        body = raw
    return json.loads(body.decode("utf-8"))


def har_entries(har: dict) -> list:
    """取出 HAR 中的所有会话条目(log.entries)。"""
    return (har.get("log") or {}).get("entries") or []


def content_to_bytes(obj: Optional[dict]) -> Optional[bytes]:
    """把 HAR 的 response.content / request.postData 对象还原为原始字节。

    HAR 规范约定:二进制正文存于 text 字段并附 encoding="base64",
    此时 text 为 base64 编码;无 encoding 时按 UTF-8 文本处理。
    """
    if not obj:
        return None
    text = obj.get("text")
    if text is None:
        return None
    if isinstance(text, bytes):
        return text
    if not isinstance(text, str):
        return None
    if obj.get("encoding") == ENCODING_BASE64:
        return base64.b64decode(text)
    return text.encode("utf-8", errors="replace")


def entry_candidate_bodies(entry: dict) -> Iterator[bytes]:
    """按优先级产出某个会话里可能携带存档的原始字节。

    优先响应体(MySekai 数据在游戏 API 的响应中),其次请求体。
    """
    response = entry.get("response") or {}
    body = content_to_bytes(response.get("content"))
    if body:
        yield body
    request = entry.get("request") or {}
    body = content_to_bytes(request.get("postData"))
    if body:
        yield body
