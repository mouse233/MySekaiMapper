"""FastAPI 上传服务。

接收分片上传的加密存档,合并后在同一进程内生成地图并推送通知
(不再通过 subprocess 调用脚本)。

同时支持 Reqable 的「上报服务器」功能:Reqable 把捕获的 HTTP 会话按
HAR 格式 POST 到 <REPORT_PATH>,本服务从会话中提取 MySekai 存档,
直接进入 解密 -> 生成地图 -> 归档 -> 推送 流水线。
"""
import asyncio
import io
import re
import secrets
import shutil
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from pathlib import Path

from fastapi import FastAPI, Request, Header, HTTPException
from fastapi.responses import PlainTextResponse

from . import config, notify as notify_mod, parser as parser_mod
from . import har as har_mod
from .render import generate as render_generate


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """uvicorn 启动完成后(INFO 日志下方)打印两个 API 端点的提示。"""

    async def _print_banner():
        await asyncio.sleep(1.0)
        print_api_banner(
            getattr(_app.state, "host", "0.0.0.0"),
            getattr(_app.state, "port", 9478),
        )

    task = asyncio.create_task(_print_banner())
    try:
        yield
    finally:
        task.cancel()


app = FastAPI(lifespan=lifespan)

# === 安全限制参数 ===
MAX_TOTAL_SIZE = 1 * 1024 * 1024    # 1MB 总大小
MAX_CHUNK_SIZE = 1 * 1024 * 1024    # 1MB 单 chunk
MAX_CHUNKS = 10                     # 最多 10 个分片
UPLOAD_ID_PATTERN = re.compile(r"^[a-zA-Z0-9_-]{1,64}$")


def _user_id_from_url(original_url):
    """从客户端上报的原始页面 URL 中提取玩家 ID。"""
    if original_url:
        m = re.search(r"/user/(\d+)", original_url)
        if m:
            return m.group(1)
    return "unknown"


def _archive_latest(user_id):
    """把 data/latest/ 复制到 archive/by-id/<user>/<时间戳>/,返回归档目录。"""
    archive_dir = (
        config.ARCHIVE_DIR / "by-id" / str(user_id)
        / datetime.now().strftime("%Y%m%d_%H%M%S")
    )
    archive_dir.mkdir(parents=True, exist_ok=True)
    for p in config.LATEST_DIR.iterdir():
        dest = archive_dir / p.name
        if p.is_dir():
            shutil.copytree(p, dest, dirs_exist_ok=True)
        else:
            shutil.copy2(p, dest)
    return archive_dir


def _generate_and_notify(bin_path, task_id, user_id):
    """同步执行:生成地图 -> 归档 -> 推送。由调用方放入线程池。"""
    render_generate(bin_path)

    image_base = None
    try:
        archive_dir = _archive_latest(user_id)
        if config.BARK_IMAGE_BASE:
            image_base = (
                config.BARK_IMAGE_BASE.rstrip("/")
                + f"/archive/by-id/{user_id}/{archive_dir.name}"
            )
    except Exception as e:
        print(f"[WARN] archive failed: {e}")

    notify_mod.notify(config.LATEST_DIR, task_id, player_id=user_id, image_base=image_base)


async def _run_generate_and_notify(bin_path, task_id, user_id):
    try:
        print(f"[LAUNCH] generating maps for {bin_path}")
        await asyncio.to_thread(_generate_and_notify, bin_path, task_id, user_id)
        print(f"[DONE] generate + notify finished for {bin_path}")
    except Exception as e:
        print(f"[ERROR] generate/notify failed: {e}")


def _save_and_launch(data: bytes, user_id: str, task_id: str):
    """把完整存档写入 RAW_DIR 并启动后台 生成->归档->推送 流水线。

    分片上传(合并后)与 Reqable 上报(整包直达)共用此入口。
    """
    output_file = config.RAW_DIR / f"mysekai_{user_id}_{task_id}.bin"
    with open(output_file, "wb") as f:
        f.write(data)
    print(f"[DONE] Mysekai saved to {output_file} ({len(data)} bytes)")

    asyncio.create_task(_run_generate_and_notify(output_file, task_id, user_id))
    print(f"[LAUNCH] background task created for {output_file}")


def _looks_like_archive(data: bytes) -> bool:
    """通过 解密+msgpack 解析 判断是否为有效的 MySekai 存档。

    上报端点会收到规则命中的全部会话(可能只是普通 API),用它过滤掉
    无关流量。AES 密钥未配置等环境问题(RuntimeError)向上抛,
    其余解析失败(数据不是存档)一律视为 False。
    """
    try:
        obj = parser_mod.decrypt_and_parse(data)
    except RuntimeError:
        raise
    except Exception:
        return False
    return isinstance(obj, dict) and "updatedResources" in obj


def print_api_banner(host: str = "0.0.0.0", port: int = 9478):
    """打印两个 API 端点的提示信息(英文)。

    REPORT_ENABLED=0 时不打印 Reqable 上报服务器那一行。
    """
    base = f"http://{host}:{port}"
    print("-" * 64)
    print(f"[API] {'Chunked upload (capture client):':<33}POST {base}/uploadMySekai")
    if config.REPORT_ENABLED:
        print(f"[API] {'Reqable report server:':<33}POST {base}{config.REPORT_PATH}")
    print("-" * 64)


@app.post("/uploadMySekai")
async def upload_chunk(
    request: Request,
    x_upload_id: str = Header(...),
    x_chunk_index: int = Header(...),
    x_total_chunks: int = Header(...),
    x_original_url: str = Header(None),
):
    # ==== upload_id 安全校验 ====
    if not UPLOAD_ID_PATTERN.fullmatch(x_upload_id):
        raise HTTPException(400, "Invalid upload id")

    # ==== chunk 参数校验 ====
    if x_total_chunks <= 0 or x_total_chunks > MAX_CHUNKS:
        raise HTTPException(400, "Invalid total chunks")

    if x_chunk_index < 0 or x_chunk_index >= x_total_chunks:
        raise HTTPException(400, "Invalid chunk index")

    data = await request.body()

    # ==== 单 chunk 大小限制 ====
    if len(data) > MAX_CHUNK_SIZE:
        raise HTTPException(413, "Chunk too large")

    upload_path = config.TMP_DIR / x_upload_id
    upload_path.mkdir(exist_ok=True)

    # ==== 总大小限制 ====
    current_size = sum(f.stat().st_size for f in upload_path.glob("chunk_*") if f.is_file())
    if current_size + len(data) > MAX_TOTAL_SIZE:
        raise HTTPException(413, "Total file too large")

    chunk_file = upload_path / f"chunk_{x_chunk_index}"
    with open(chunk_file, "wb") as f:
        f.write(data)

    print(f"[UPLOAD] {x_upload_id} chunk {x_chunk_index+1}/{x_total_chunks} ({len(data)} bytes)")

    existing_chunks = list(upload_path.glob("chunk_*"))
    if len(existing_chunks) == x_total_chunks:
        print(f"[MERGE] {x_upload_id} all chunks received, merging...")

        # 合并前兜底总大小校验
        total_size = sum(f.stat().st_size for f in existing_chunks)
        if total_size > MAX_TOTAL_SIZE:
            shutil.rmtree(upload_path, ignore_errors=True)
            raise HTTPException(413, "Merged file too large")

        user_id = _user_id_from_url(x_original_url)

        merged = io.BytesIO()
        for i in range(x_total_chunks):
            chunk_path = upload_path / f"chunk_{i}"
            with open(chunk_path, "rb") as infile:
                shutil.copyfileobj(infile, merged)

        shutil.rmtree(upload_path, ignore_errors=True)
        _save_and_launch(merged.getvalue(), user_id, x_upload_id)

    return PlainTextResponse("OK")


@app.post(config.REPORT_PATH)
async def reqable_report(request: Request):
    """Reqable 上报服务器端点:接收 HAR JSON,提取 MySekai 存档进入流水线。

    - 请求体为 HAR 格式 JSON,支持 Content-Encoding: gzip / br / zstd
    - Reqable 每个会话只上报 1 次且失败不重试,因此这里尽快返回 200
    - 每次请求最多处理 1 份存档(取第一个有效条目),避免重复推送
    """
    if not config.REPORT_ENABLED:
        raise HTTPException(404, "Report server disabled")

    if config.REPORT_TOKEN:
        token = request.headers.get("X-Report-Token", "")
        if not secrets.compare_digest(token, config.REPORT_TOKEN):
            raise HTTPException(401, "Invalid report token")

    raw = await request.body()
    if len(raw) > config.REPORT_MAX_SIZE:
        raise HTTPException(413, "Report body too large")

    content_encoding = request.headers.get("content-encoding")
    try:
        har = har_mod.parse_har(raw, content_encoding)
    except Exception as e:
        print(
            f"[REPORT] invalid HAR body: {e} "
            f"(Content-Encoding={content_encoding!r}, head={raw[:8].hex()})"
        )
        raise HTTPException(400, "Invalid HAR body")

    platform = request.headers.get("x-reqable-platform", "-")
    host = request.headers.get("x-reqable-reporter-host", "-")
    rule = request.headers.get("x-reqable-reporter-rule", "-")
    entries = har_mod.har_entries(har)

    processed = 0
    skipped = 0
    for entry in entries:
        url = ((entry.get("request") or {}).get("url")) or ""
        user_id = _user_id_from_url(url)
        for data in har_mod.entry_candidate_bodies(entry):
            if len(data) > MAX_TOTAL_SIZE:
                skipped += 1
                continue
            try:
                valid = _looks_like_archive(data)
            except RuntimeError as e:
                print(f"[REPORT] cannot validate archive: {e}")
                raise HTTPException(500, "AES keys not configured")
            if not valid:
                skipped += 1
                continue
            _save_and_launch(data, user_id, uuid.uuid4().hex[:12])
            processed += 1
            break
        if processed:
            break

    print(
        f"[REPORT] platform={platform} host={host} rule={rule} "
        f"entries={len(entries)} processed={processed} skipped={skipped}"
    )
    return PlainTextResponse("ok")


if __name__ == "__main__":
    import uvicorn
    app.state.host = "0.0.0.0"
    app.state.port = 9478
    uvicorn.run(app, host="0.0.0.0", port=9478)
