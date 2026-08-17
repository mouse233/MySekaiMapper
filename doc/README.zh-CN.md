# MySekaiMapper

🌐 语言: [English](../README.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/zh-CN/>

《世界计划 多彩舞台》（Project Sekai）MySekai（我的世界）采集点地图生成工具。

**项目初衷**：搭配 MitM 模块或 Reqable 的「上报服务器」功能使用——抓包工具捕获游戏内 MySekai 数据包后，自动上传到本服务（一次 POST 即可，分片上传亦受支持）；服务端解密加密存档、提取各站点的资源掉落坐标，绘制采集地图，再把结果（含稀有资源统计）推送到玩家的 Telegram / Bark（iOS Day.app），全程无需人工介入。

一次任务会生成 **4 张地图**：`site_5.png`（初始空地）、`site_6.png`（心愿沙滩）、`site_7.png`（烂漫花田）、`site_8.png`（忘却之所），外加一份 `rare_resources.txt` 稀有资源统计。

本项目已在朝夕光年（Nuverse）运营的 CN 服 / TW 服中测试通过，其他服务器可用性未知。

## 工作流程

```
游戏 API 响应 → MitM 模块 / Reqable 上报服务器（抓包捕获 mysekai 数据）
   │  ① 自动上传（一次 POST，分片亦支持）→ server.py 自动处理
   │  ② 或手动放置 .bin 存档 → cli.py generate
   ▼
parser.py    AES-128-CBC 解密 + msgpack 解析 + 坐标旋转
   ▼
render.py    绘制 site_5.png ~ site_8.png + rare_resources.txt → data/latest/
   ▼
notify.py    推送：
             ├─ Telegram  ：图片 multipart 直传，无需公网直链 ← 默认渠道
             └─ Bark      ：以 image= URL 直链通知，需静态文件服务器
```

## 快速上手

先完成安装与 `.env` 基础配置，再按你想要的推送方式选择路径：

- **路径 A（仅 Telegram Bot 推送）**：配置最少，推荐先跑通这条；
- **路径 B（启用 Bark 推送）**：在路径 A 基础上，需要额外配置 Bark key、玩家路由与静态文件服务器。

### 1. 安装

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 可选:安装 mysekai 命令(等价于 python cli.py ...)
venv/bin/pip install -e .
```

### 2. 配置 .env（必填项）

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV` 为 MySekai 存档的 AES-128-CBC 解密密钥（各 16 字节），无论走哪条路径都必须填写。其余变量按选择的路径配置：

| 变量 | 必填 | 说明 |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | MySekai 存档的 AES-128-CBC 密钥，各 16 字节 |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | 选* | Telegram 推送（默认渠道）需要，来自 [@BotFather](https://t.me/BotFather) |
| `BARK_ICON` | 选 | Bark 通知图标 URL |
| `BARK_IMAGE_BASE` | 选 | 静态文件服务器根地址（推送 Bark 图片直链用，见下文） |
| `FALLBACK_IMAGE_BASE` | 选 | 未配置 `BARK_IMAGE_BASE` 时的图片直链兜底地址 |

> \* 若只想用 Bark 收通知：可留空 Telegram 配置，但**必须在 `config/push_map.json` 里把玩家路由到 Bark 别名**，否则未配置玩家默认走 Telegram，而 Telegram 缺配置时只会打印一行警告并跳过，结果是什么都不推。

### 3. 路径 A：仅 Telegram Bot 推送（最简）

适用场景：只要在 Telegram 收到地图与统计，不折腾其他组件。

1. 在 `.env` 中填写 Telegram 配置（来自 [@BotFather](https://t.me/BotFather)）：

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 手动跑一遍验证：

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. 日常使用：启动上传服务，存档到达后自动生成地图并推送。两种抓包方式任选：

   - **MitM 模块**：按「上传接口」上传存档
   - **Reqable 上报服务器**：配置匹配规则与上报路径（见下文「Reqable 上报服务器」章节）

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

路径 A **不需要**：`config/push_map.json`、`config/bark_map.json`、静态文件服务器、`BARK_IMAGE_BASE`。未配置的玩家默认就推送到 Telegram。

### 4. 路径 B：启用 Bark 推送（需额外配置）

在路径 A 的基础上（Telegram 配置可保留，也可留空只推 Bark），按顺序补齐：

1. **配置 Bark key**：在 `config/bark_map.json` 中为每个别名配置设备 key（模板见同目录 `bark_map.example.json`）。
2. **配置玩家路由**：在 `config/push_map.json` 中把玩家 ID 路由到 Bark 别名，例如：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ⚠️ **必须配置**：未配置的玩家默认走 Telegram；若此时 Telegram 又未配置，只会打印警告并跳过，结果什么都不推。
3. **搭建静态文件服务器**：把项目的 `data/` 目录暴露为公网可达的 HTTP(S) 服务，并在 `.env` 设置 `BARK_IMAGE_BASE=https://<域名或IP:端口>`。否则 Bark 通知不带地图图片（详见下文「搭建静态文件服务器」）。
4. 验证与日常使用同路径 A（第 2、3 步）。

## 上传接口

客户端把捕获的 mysekai 响应体通过 `POST /uploadMySekai` 上传（一次 POST 即可；分片上传仅作兼容保留）。手动用 curl 按同一协议调试亦可。header 如下：

| Header | 说明 |
| --- | --- |
| `X-Upload-Id` | 上传任务 ID（仅字母数字与 `-` / `_`，长度 1~64），必填 |
| `X-Chunk-Index` | 分片序号，从 0 开始（单片上传恒为 0），必填 |
| `X-Total-Chunks` | 总分片数（1~10；单片上传填 1），必填 |
| `X-Original-Url` | 客户端原始页面 URL，用于解析玩家 ID（如 `https://.../user/123456...`）；**可选**，缺失时玩家 ID 记为 `unknown` |
| `X-Script-Version` | 客户端脚本版本号；服务端忽略该头，可不传 |

请求体为原始二进制存档数据（无需 multipart）。

限制：

- 单文件总大小 ≤1MB（`MAX_TOTAL_SIZE`）
- 单个分片 ≤1MB（`MAX_CHUNK_SIZE`，超限返回 413）
- 总分片数 ≤10（`MAX_CHUNKS`）

> 注意：当前存档约 200KB，**一次 POST 即可传完**。分片上传仅为兼容旧抓包客户端保留；若使用分片，每片应明显小于 1MB（例如 256KB），10 片可传满 1MB 上限。

响应：

| 状态码 | 含义 |
| --- | --- |
| `200` | 存档已接收，返回 `OK`；服务端自动完成：合并存档（如分片）→ 生成地图 → 归档到 `data/archive/by-id/<user_id>/<时间戳>/` → 推送通知，全程无需人工介入 |
| `400` | 参数非法（upload id 格式错误、分片序号越界、总分片数不在 1~10） |
| `413` | 超过大小限制（单分片超 1MB，或累计总大小超 1MB） |

### curl 示例

单次 POST（当前存档一次即可传完）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

## Reqable 上报服务器

不依赖自定义抓包客户端，也可以直接用 Reqable 内置的「上报服务器」功能（Reqable v2.20.0+）：它会把每个已捕获的 HTTP 会话按 [HAR](https://en.wikipedia.org/wiki/HAR_(file_format)) JSON 格式自动 POST 到你的服务器，可选 gzip / brotli / zstd 压缩。上报端点**默认开启**，与分片上传共存——`python cli.py server` 同时提供两个接口；设 `REPORT_ENABLED=0` 可关闭：

```bash
python cli.py server
```

配置（`.env`）：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1`（开启） | 设 `0` / `false` 关闭上报端点 |
| `REPORT_PATH` | `/reqable/report` | 端点路径，填入 Reqable 的「上报路径」 |
| `REPORT_MAX_SIZE` | `1` | HAR 请求体大小上限（MB，默认 1，与分片上传上限一致） |
| `REPORT_TOKEN` | （空） | 可选共享令牌；设置后端点要求请求头 `X-Report-Token` 匹配 |

每次上报，服务端会：

1. 按 `Content-Encoding`（gzip / br / zstd）解压并解析 HAR。
2. 遍历 `log.entries`，取第一个「响应体（兜底：请求体）能用 `AES_KEY` / `AES_IV` 解密并解析为 MySekai 存档」的会话——命中规则但与存档无关的流量会被跳过。
3. 从会话 URL 解析玩家 ID（`/user/<id>`，与 `X-Original-Url` 同规则）。
4. 存档保存到 `data/raw_mysekai/`，并启动与分片上传相同的 生成 → 归档 → 推送 流水线。

注意：

- Reqable 每个会话**只上报 1 次且失败不重试**，因此端点会尽快返回 `200`。请保持服务稳定，并留意 `[REPORT]` 日志。
- 每次上报只处理 **1 份**存档（第一个有效条目），因此匹配多个接口的规则不会造成重复推送。
- 安全：协议本身没有鉴权。Reqable 无法附加自定义请求头，建议把随机串拼进 `REPORT_PATH`（如 `/reqable/report/9f3a…`），或用反向代理 / 防火墙做访问限制，而不是依赖 `REPORT_TOKEN`。

Reqable 侧配置示例：

- URL 匹配规则：`https://<游戏API域名>/api/user/*/mysekai*`
- 上报路径：`http://<你的服务器>:9478/reqable/report`
- 压缩算法：gzip / brotli / zstd 均可（服务端三种都支持）

五个服务器的游戏 API 域名：

| 服务器 | 游戏 API 域名 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

推荐匹配规则：`https://<域名>/api/user/*/mysekai*`（CN 已实测验证）。若你所在服务器的 mysekai 接口路径不同，请按实际路径调整规则。

手动 curl 验证（gzip 压缩的 HAR）：

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```

## 推送机制

### 默认走 Telegram Bot

- 未在 `config/push_map.json` 中配置的玩家，**一律默认推送到 Telegram**；`push_map.json` 文件缺失时同样默认 Telegram。
- Telegram 使用 Bot API `sendMediaGroup`，把 4 张本地 PNG 作为 multipart 直接上传，**不需要公网直链，也不依赖静态文件服务器**；`TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` 缺失时只打印警告并跳过，不影响 Bark 渠道。

### Bark 推送依赖公开直链

Bark（Day.app）通知中的图片是 **URL 直链**：`notify.py` 把图片地址编码进 `image=` 参数发给 `api.day.app`，由 Bark 服务器再去抓取这张图。因此该 URL 必须**公网可达（建议 HTTPS）**，否则 Bark 通知里没有图片。

4 张地图的直链由 `notify.py` 按以下优先级拼出：

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| 场景 | base 取值 | 图片直链形态 |
| --- | --- | --- |
| 服务器流程（推荐） | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<时间戳>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<时间戳>/site_{5..8}.png` |
| 手动 CLI 推送 | `BARK_IMAGE_BASE` 或 `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`（需把 `data/latest/` 暴露在 `<base>/` 下） |

> 注意：服务器流程只有在配置了 `BARK_IMAGE_BASE` 时才会拼出带归档路径的直链；若只配了 `FALLBACK_IMAGE_BASE`，服务器推送的直链同样是 `<FALLBACK_IMAGE_BASE>/site_{5..8}.png`。

## 静态文件服务器示例（可选）

目的：把 `data/archive/` 目录暴露成公开 URL，让 Bark 服务器能抓到四张地图。

**推荐做法**：静态服务器的根目录指向项目的 `data/`，再设置 `BARK_IMAGE_BASE=https://<你的域名或IP:端口>`，即可自动映射：

```
data/archive/by-id/<user_id>/<时间戳>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<时间戳>/site_5.png
```

常用示例：

Python 内置（最简，适合内网/测试）：

```bash
python -m http.server 8000 --directory data
# 然后设置 BARK_IMAGE_BASE=http://<服务器IP>:8000
```

nginx：

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    # ... ssl 证书配置 ...
    root /path/to/MySekaiMapper/data;
}
```

Caddy（自动 HTTPS）：

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

注意事项：

- **不要用 `127.0.0.1` / `localhost`** 作为直链地址；Bark 服务器需要能访问该地址，一般直接选公网可达的地址，内网 IP 仅在确认互通时使用。
- **只用 Telegram 则完全不需要静态服务器**，跳过本节即可。
- 手动 `cli.py notify` 的直链不带归档路径，需要另把 `data/latest/` 暴露在 `BARK_IMAGE_BASE` 下；或用 `FALLBACK_IMAGE_BASE` 指向输出目录（例如 `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → 该服务器把 `data/latest/` 挂在 `/output` 下）。

## 玩家推送路由（可选）

在 `config/` 下按需创建本地配置（格式见同目录 `*.example.json`，已被 `.gitignore` 忽略）：

- `push_map.json` — 玩家 ID → 推送方式：值为 `"telegram"`、Bark 别名、`"none"`（不推送），也支持组合写法 `["alias", "telegram"]` 或 `"alias+tg"`。**未配置的玩家默认 `telegram`**。

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark 别名 → 设备 key：

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 常见问题

- **Bark 通知收不到图片？** 检查直链是否公网可达：在浏览器/手机网络下直接打开 `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<时间戳>/site_5.png` 应能显示图片；内网地址、`127.0.0.1`、或证书异常的 HTTPS 都会导致抓图失败。
- **什么都没推送？** 检查 `push_map.json` 是否把该玩家设成了 `"none"`；只配了 Bark 的用户是否忘了在该玩家上配置 Bark 别名（未配置玩家默认走 Telegram）；Telegram 渠道是否配了 token 与 chat id；Bark 渠道是否缺 key（报 `[BARK] ... failed` 日志）。
- **不想收到 Bark 只想要 Telegram？** 什么都不用做——未配置的玩家默认就走 Telegram。

## 命令行工具（cli.py）

所有功能都可通过 `cli.py` 驱动；安装后（`pip install -e .`）也可用等价的 `mysekai` 命令。命令成功退出码为 0，出错为 1（错误信息打印到 stderr）。

```bash
python cli.py --help           # 子命令总览
python cli.py <命令> --help     # 查看某子命令的参数
```

### generate —— 解密存档并生成地图

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`：加密存档路径（.bin），必填
- 流程：AES-128-CBC 解密 → msgpack 解析 → 提取掉落坐标 → 绘制 4 张地图（`site_5.png` ~ `site_8.png`）→ 写出 `rare_resources.txt`
- 输出到 `data/latest/`，结束时打印实际路径
- 前置要求：`.env` 已配置 `AES_KEY` / `AES_IV`；存档中没有任何掉落点时会报错退出

### notify —— 推送地图与统计

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`：包含 `site_*.png` 与 `rare_resources.txt` 的目录（通常就是 `data/latest/`）
- `[task_id]`：可选，上传任务 ID，默认 `unknown`。用于从 `data/raw_mysekai/` 反查玩家 ID：优先匹配 `mysekai_<玩家ID>_<task_id>.bin`，匹配不到时取 raw_mysekai 里最新的存档
- 推送到 Telegram 还是 Bark 由 `config/push_map.json` 路由（未配置的玩家默认走 Telegram），详见「玩家推送路由」

### server —— 启动上传服务（分片上传 + Reqable 上报服务器）

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- 启动 FastAPI 服务：客户端向 `POST /uploadMySekai` 上传加密存档（单片或分片；接口细节见「上传接口」）；Reqable 也可把 HAR 会话上报到内置上报端点（见上文「Reqable 上报服务器」章节）
- 全部片到达后自动完成：合并存档 → 生成地图 → 归档到 `data/archive/by-id/<user_id>/<时间戳>/` → 按玩家路由推送通知，无需人工介入
- 默认监听 `9478` 端口；公网部署时建议通过反向代理暴露为 HTTPS，客户端脚本中写死的上传 URL（含端口）需与你的实际部署保持一致

### 典型手动流程

```bash
python cli.py generate mysekai_xxx.bin       # 1. 生成地图到 data/latest/
python cli.py notify data/latest <task_id>   # 2. 推送（task_id 填上传 ID，如 chfto53c3）
```

## 目录结构

```
├── app/                       # 核心包
│   ├── config.py              # 路径／环境变量／本地配置集中管理
│   ├── crypto.py              # MySekai 存档 AES-128-CBC 解密
│   ├── parser.py              # msgpack 解析＋站点坐标旋转（纯函数）
│   ├── render.py              # 提取掉落点 → matplotlib 绘图＋稀有资源统计
│   ├── notify.py              # 推送：Telegram 媒体组／Bark，按玩家路由
│   ├── server.py              # FastAPI 分片上传服务
│   └── cli.py                 # 命令行入口
├── assets/                    # 静态资源（提交到仓库）
│   ├── resourceId.csv         # 物品 ID → 名称＋图标（base64）
│   └── NotoSansSC-Regular.ttf # 中文字体（OFL 协议）
├── config/                    # 本地配置（真实文件不提交，参考 *.example.json）
│   ├── bark_map.example.json  # Bark 别名 → 设备 key 模板
│   └── push_map.example.json  # 玩家 ID → 推送方式模板
├── data/                      # 运行时数据（整个目录 gitignore）
│   ├── tmp/                   # 分片上传暂存，合并后即清
│   ├── raw_mysekai/           # 合并后的原始（加密）存档，永久保留
│   ├── archive/               # 历史成品归档 by-id/<user>/<时间戳>/（Bark 直链即指向此处）
│   └── latest/                # 最近一次生成的成品
├── cli.py                     # 统一入口
├── tests/                     # 单元测试（pytest）
├── .env.example               # 环境变量模板（复制为 .env 填写）
└── requirements.txt           # 运行时依赖（精确锁版本）
```

## 测试

```bash
python -m pytest
```

## 免责声明

本工具仅用于个人学习与娱乐，请勿用于任何商业用途或违反游戏服务条款的行为。游戏数据与美术资源版权归原版权方所有。

## 许可证

本项目代码采用 [MIT License](LICENSE)（版权所有 © 2025 mouse233），可自由使用、修改与再分发，详见 [LICENSE](LICENSE)。

> ⚠️ 许可证仅覆盖本项目代码：`assets/` 中的游戏素材（如 `resourceId.csv` 内的物品图标）与游戏数据版权归 SEGA / Colorful Palette 等原版权方所有，**不在 MIT 授权范围内**，请勿将其用于本工具之外的用途。
