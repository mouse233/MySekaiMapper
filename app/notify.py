"""推送:Telegram 与 Bark(iOS Day.app)。

- Telegram:地图图片按媒体组发送,附稀有资源统计摘要
- Bark:按玩家配置路由到不同的设备 key
"""
import json
import re
import sys
from pathlib import Path

import requests
from urllib.parse import quote_plus

from . import config


# ================== 本地配置读取 ==================

def load_bark_map():
    """dict alias -> bark key(文件缺失/损坏时返回空 dict)。"""
    if not config.BARK_MAP_FILE.exists():
        return {}
    try:
        return json.loads(config.BARK_MAP_FILE.read_text(encoding="utf-8")) or {}
    except Exception:
        return {}


def load_push_map():
    """dict player_id -> 推送方式。"""
    if not config.PUSH_MAP_FILE.exists():
        return {}
    try:
        return json.loads(config.PUSH_MAP_FILE.read_text(encoding="utf-8")) or {}
    except Exception:
        return {}


def get_bark_key_for(alias=None, explicit_key=None):
    """解析该用哪个 Bark key,优先级:

    1. explicit_key(直接传入)
    2. alias -> bark_map.json

    返回 key 字符串或 None(未命中)。
    """
    if explicit_key:
        return explicit_key
    if alias:
        return load_bark_map().get(alias) or None
    return None


# ================== Bark 推送 ==================

def bark_send(title, body=None, image_url=None, alias=None,
              explicit_key=None, icon_url=None, timeout=10):
    """发送一条 Bark 通知,返回 (ok: bool, message: str)。"""
    key = get_bark_key_for(alias=alias, explicit_key=explicit_key)
    if not key:
        return False, "BARK key not configured"
    enc_title = quote_plus(title)
    base = f"https://api.day.app/{key}/{enc_title}"
    params = []
    if body:
        params.append("body=" + quote_plus(str(body)))
    if image_url:
        params.append("image=" + quote_plus(image_url))
    if icon_url:
        params.append("icon=" + quote_plus(icon_url))
    url = base + ("?" + "&".join(params) if params else "")
    try:
        r = requests.get(url, timeout=timeout)
        return r.status_code // 100 == 2, f"Status {r.status_code}"
    except Exception as e:
        return False, str(e)


# ================== Telegram 推送 ==================

def tg_send_text(text):
    if not config.TELEGRAM_BOT_TOKEN or not config.TELEGRAM_CHAT_ID:
        print("[TG] TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not set")
        return
    url = f"https://api.telegram.org/bot{config.TELEGRAM_BOT_TOKEN}/sendMessage"
    data = {"chat_id": config.TELEGRAM_CHAT_ID, "text": text}
    try:
        requests.post(url, data=data, timeout=10)
    except Exception as e:
        print("[TG] send text failed:", e)


def tg_send_media_group(image_paths, caption=None):
    """把图片按媒体组发送(每组最多 10 张),说明文字挂在第一张上。"""
    if not config.TELEGRAM_BOT_TOKEN or not config.TELEGRAM_CHAT_ID:
        print("[TG] TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not set")
        return

    def send_batch(batch, first_caption=None):
        if len(batch) == 1:
            url = f"https://api.telegram.org/bot{config.TELEGRAM_BOT_TOKEN}/sendPhoto"
            files = {"photo": open(batch[0], "rb")}
            data = {"chat_id": config.TELEGRAM_CHAT_ID}
            if first_caption:
                data["caption"] = first_caption[:1000]
            try:
                requests.post(url, data=data, files=files, timeout=30)
            except Exception as e:
                print(f"[TG] sendPhoto failed: {e}")
            finally:
                files["photo"].close()
            return

        url = f"https://api.telegram.org/bot{config.TELEGRAM_BOT_TOKEN}/sendMediaGroup"
        media = []
        files = {}
        for i, p in enumerate(batch):
            attach_name = f"photo{i}"
            item = {"type": "photo", "media": f"attach://{attach_name}"}
            if i == 0 and first_caption:
                item["caption"] = first_caption[:1000]
            media.append(item)
            files[attach_name] = (Path(p).name, open(p, "rb"))

        data = {"chat_id": config.TELEGRAM_CHAT_ID, "media": json.dumps(media)}
        try:
            requests.post(url, data=data, files=files, timeout=60)
        except Exception as e:
            print(f"[TG] sendMediaGroup failed: {e}")
        finally:
            for f in files.values():
                try:
                    f[1].close()
                except Exception:
                    pass

    imgs = list(image_paths)
    batch_size = 10
    for idx in range(0, len(imgs), batch_size):
        batch = imgs[idx: idx + batch_size]
        first_caption = caption if idx == 0 else None
        send_batch(batch, first_caption)


# ================== 玩家识别 ==================

def detect_player_id(task_id):
    """从 data/raw_mysekai/ 反查 task_id 对应的玩家 ID(兼容未显式传入的调用)。"""
    try:
        candidates = []
        if task_id and task_id != "unknown":
            candidates = list(config.RAW_DIR.glob(f"*_{task_id}.bin"))
        if not candidates:
            candidates = list(config.RAW_DIR.glob("mysekai_*.bin"))
            candidates.sort(key=lambda p: p.stat().st_mtime, reverse=True)
        if candidates:
            m = re.search(r"mysekai_(\d+)_", candidates[0].name)
            if m:
                return m.group(1)
    except Exception as e:
        print(f"[WARN] failed to detect player id: {e}")
    return None


# ================== 推送编排 ==================

def resolve_method(m):
    """把推送配置解析为 (bark_aliases: list, send_telegram: bool)。

    兼容旧版字符串与新版 list 两种写法:
    - 'none' / []            -> 不推送
    - 'telegram'             -> 仅 Telegram
    - 'klee+tg'              -> 别名 klee + Telegram
    - ['telegram', 'dodoco'] -> list 写法
    """
    if not m:
        return [], False
    sel = []
    if isinstance(m, list):
        sel = m
    else:
        if m == 'none':
            sel = []
        elif m == 'telegram':
            sel = ['telegram']
        elif '+tg' in str(m):
            alias = str(m).split('+tg')[0]
            sel = [alias, 'telegram']
        else:
            sel = [str(m)]

    send_telegram = 'telegram' in sel
    bark_aliases = [s for s in sel if s != 'telegram']
    return bark_aliases, send_telegram


def notify(output_dir, task_id="unknown", player_id=None, image_base=None):
    """推送地图与统计。

    Args:
        output_dir: 包含 site_*.png 与 rare_resources.txt 的目录(通常为 data/latest/)
        task_id: 任务标识(上传 ID)
        player_id: 玩家 ID;为 None 时自动从 raw_mysekai/ 反查
        image_base: 图片公开 URL 根路径,优先于环境变量 BARK_IMAGE_BASE
    """
    output_dir = Path(output_dir)

    rare_file = output_dir / "rare_resources.txt"
    if rare_file.exists():
        text = rare_file.read_text(encoding="utf-8")
    else:
        text = "Mysekai 抓取完成,但未生成稀有资源统计。"

    if player_id is None:
        player_id = detect_player_id(task_id)

    header = f"🎮 Mysekai 抓取完成\nTask: {task_id}\n"
    if player_id:
        header += f"Player: {player_id}\n"
    header += "\n"
    full_text = header + text

    images = sorted(output_dir.glob("site_*.png"))

    # 按玩家选择推送方式;未配置的玩家默认走 Telegram
    method = load_push_map().get(str(player_id)) if player_id else None
    if not method:
        method = 'telegram'

    # 稀有资源统计(压缩成一行,先推)
    rare_lines = [ln.strip() for ln in text.splitlines() if ln.strip()]
    rare_compact = " | ".join(rare_lines[:8]) if rare_lines else "稀有资源统计: 无"

    bark_aliases, send_telegram = resolve_method(method)

    if bark_aliases:
        for alias in bark_aliases:
            ok, msg = bark_send(rare_compact, alias=alias,
                                icon_url=config.BARK_ICON)
            if not ok:
                print(f"[BARK] rare push failed for {alias}: {msg}")

    # 依次推送 site5..site8 地图
    base = image_base or config.BARK_IMAGE_BASE or config.FALLBACK_IMAGE_BASE
    if bark_aliases:
        for i in range(5, 9):
            title = f"site{i}"
            image_url = base.rstrip('/') + f"/site_{i}.png" if base else None
            for alias in bark_aliases:
                ok, msg = bark_send(title, image_url=image_url, alias=alias,
                                    icon_url=config.BARK_ICON)
                if not ok:
                    print(f"[BARK] image push failed for site{i} -> {alias}: {msg}")

    if send_telegram:
        if images:
            tg_send_media_group(images, caption=full_text)
        else:
            tg_send_text(full_text)
    else:
        print(f"[INFO] skipping Telegram push for player {player_id} (method={method})")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python -m app.notify <output_dir> [task_id]")
        sys.exit(1)
    out_dir = Path(sys.argv[1])
    task = sys.argv[2] if len(sys.argv) > 2 else "unknown"
    notify(out_dir, task)
