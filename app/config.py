"""集中管理路径、环境变量与本地配置。

所有模块统一从这里取路径/密钥,不再各自用 Path(__file__).parent 推算。
"""
import os
from pathlib import Path

from dotenv import load_dotenv

# 项目根目录(本文件位于 <root>/app/config.py)
BASE_DIR = Path(__file__).resolve().parent.parent

# 目录布局(支持环境变量覆盖,便于部署到其他位置)
ASSETS_DIR = Path(os.environ.get("MYSK_ASSETS_DIR", BASE_DIR / "assets"))
CONFIG_DIR = Path(os.environ.get("MYSK_CONFIG_DIR", BASE_DIR / "config"))
DATA_DIR = Path(os.environ.get("MYSK_DATA_DIR", BASE_DIR / "data"))

load_dotenv(BASE_DIR / ".env")

# ---- 静态资源(提交到仓库) ----
RESOURCE_CSV = ASSETS_DIR / "resourceId.csv"
FONT_FILE = ASSETS_DIR / "NotoSansSC-Regular.ttf"

# ---- 运行时数据目录(全部 gitignore,首次导入自动创建) ----
TMP_DIR = DATA_DIR / "tmp"           # 分片上传暂存,合并后即清
RAW_DIR = DATA_DIR / "raw_mysekai"   # 合并后的原始(加密)存档,永久保留
ARCHIVE_DIR = DATA_DIR / "archive"   # 历史成品归档 by-id/<user>/<时间戳>/
LATEST_DIR = DATA_DIR / "latest"     # 最近一次生成的成品

for _d in (TMP_DIR, RAW_DIR, ARCHIVE_DIR, LATEST_DIR):
    _d.mkdir(parents=True, exist_ok=True)

# ---- 通知配置(从 .env 读取) ----
TELEGRAM_BOT_TOKEN = os.environ.get("TELEGRAM_BOT_TOKEN")
TELEGRAM_CHAT_ID = os.environ.get("TELEGRAM_CHAT_ID")
BARK_ICON = os.environ.get("BARK_ICON")
BARK_IMAGE_BASE = os.environ.get("BARK_IMAGE_BASE")
FALLBACK_IMAGE_BASE = os.environ.get("FALLBACK_IMAGE_BASE")

# ---- Reqable 上报服务器(可选,默认关闭) ----
# 开启后服务端额外提供 POST <REPORT_PATH> 端点,接收 Reqable「上报服务器」
# 功能按 HAR 格式上报的会话数据(支持 gzip / brotli / zstd 压缩)。
REPORT_ENABLED = os.environ.get("REPORT_ENABLED", "").strip().lower() in ("1", "true", "yes", "on")
REPORT_PATH = os.environ.get("REPORT_PATH", "/reqable/report")
# HAR JSON 请求体大小上限(默认 8MB;存档 ≤1MB,base64 膨胀约 33%,留足余量)
REPORT_MAX_SIZE = int(os.environ.get("REPORT_MAX_SIZE", "8")) * 1024 * 1024
# 可选共享令牌:设置后上报端点要求请求头 X-Report-Token 与之匹配
# (Reqable 本身无法附加自定义请求头,若用它上报建议把随机串拼进 REPORT_PATH 代替,
#  或依靠反向代理 IP 白名单防护)
REPORT_TOKEN = os.environ.get("REPORT_TOKEN", "")

# ---- 本地 JSON 配置(含设备密钥/玩家 ID,不提交,参考 config/*.example.json) ----
BARK_MAP_FILE = CONFIG_DIR / "bark_map.json"
PUSH_MAP_FILE = CONFIG_DIR / "push_map.json"

# AES 密钥在 app/crypto.py 中按需从环境读取并校验(便于测试注入)
