# MySekaiMapper
🌐 語言: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/zh-TW/>

《世界計畫 繽紛舞台！feat. 初音未來》（Project Sekai）MySekai（我的世界）採集點地圖產生工具。

**專案初衷**：搭配 MitM 模組或 Reqable 的「上報伺服器」功能使用——抓封包工具擷取遊戲內 MySekai 資料封包後，自動分片上傳到本服務；伺服器端合併加密存檔、解密並擷取各站點的資源掉落座標，繪製採集地圖，再把結果（含稀有資源統計）推播到玩家的 Telegram / Bark（iOS Day.app），全程無需人工介入。

一次任務會產生 **4 張地圖**：`site_5.png`（初始空地）、`site_6.png`（心願沙灘）、`site_7.png`（爛漫花田）、`site_8.png`（忘卻之所），外加一份 `rare_resources.txt` 稀有資源統計。

本專案已在朝夕光年（Nuverse）營運的 CN 服 / TW 服中測試通過，其他伺服器可用性未知。

## 工作流程

```
遊戲 API 回應 → MitM 模組 / Reqable 上報伺服器（抓封包擷取 mysekai 資料）
   │  ① 自動分片上傳 → server.py 自動合併（推薦，專案初衷）
   │  ② 或手動放置 .bin 存檔 → cli.py generate
   ▼
parser.py    AES-128-CBC 解密 + msgpack 解析 + 座標旋轉
   ▼
render.py    繪製 site_5.png ~ site_8.png + rare_resources.txt → data/latest/
   ▼
notify.py    推播：
             ├─ Telegram  ：圖片 multipart 直傳，無需公開網路直連 ← 預設管道
             └─ Bark      ：以 image= URL 直連通知，需靜態檔案伺服器
```

## 快速上手

先完成安裝與 `.env` 基礎設定，再依你想要的推播方式選擇路徑：

- **路徑 A（僅 Telegram Bot 推播）**：設定最少，建議先跑通這條；
- **路徑 B（啟用 Bark 推播）**：在路徑 A 的基礎上，需要額外設定 Bark key、玩家路由與靜態檔案伺服器。

### 1. 安裝

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 可選:安裝 mysekai 命令(等同於 python cli.py ...)
venv/bin/pip install -e .
```

### 2. 設定 .env（必填項目）

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV` 為 MySekai 存檔的 AES-128-CBC 解密金鑰（各 16 位元組），無論走哪條路徑都必須填寫。其餘變數依選擇的路徑設定：

| 變數 | 必填 | 說明 |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | MySekai 存檔的 AES-128-CBC 金鑰，各 16 位元組 |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | 選* | Telegram 推播（預設管道）需要，來自 [@BotFather](https://t.me/BotFather) |
| `BARK_ICON` | 選 | Bark 通知圖示 URL |
| `BARK_IMAGE_BASE` | 選 | 靜態檔案伺服器根位址（推播 Bark 圖片直連用，見下文） |
| `FALLBACK_IMAGE_BASE` | 選 | 未設定 `BARK_IMAGE_BASE` 時的圖片直連備用位址 |

> \* 若只想用 Bark 收通知：可留空 Telegram 設定，但**必須在 `config/push_map.json` 裡把玩家路由到 Bark 別名**，否則未設定的玩家預設走 Telegram，而 Telegram 缺設定時只會印出一行警告並跳過，結果是什麼都不推。

### 3. 路徑 A：僅 Telegram Bot 推播（最簡）

適用場景：只要在 Telegram 收到地圖與統計，不折騰其他元件。

1. 在 `.env` 中填寫 Telegram 設定（來自 [@BotFather](https://t.me/BotFather)）：

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 手動跑一遍驗證：

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. 日常使用：啟動上傳服務；抓封包用戶端（MitM 模組 / Reqable 上報伺服器）按「上傳介面」分片上傳後，自動產生地圖並推播：

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

路徑 A **不需要**：`config/push_map.json`、`config/bark_map.json`、靜態檔案伺服器、`BARK_IMAGE_BASE`。未設定的玩家預設就推播到 Telegram。

### 4. 路徑 B：啟用 Bark 推播（需額外設定）

在路徑 A 的基礎上（Telegram 設定可保留，也可留空只推 Bark），依順序補齊：

1. **設定 Bark key**：在 `config/bark_map.json` 中為每個別名設定裝置 key（範本見同目錄 `bark_map.example.json`）。
2. **設定玩家路由**：在 `config/push_map.json` 中把玩家 ID 路由到 Bark 別名，例如：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ⚠️ **必須設定**：未設定的玩家預設走 Telegram；若此時 Telegram 又未設定，只會印出警告並跳過，結果什麼都不推。
3. **架設靜態檔案伺服器**：把專案的 `data/` 目錄暴露為公開網路可達的 HTTP(S) 服務，並在 `.env` 設定 `BARK_IMAGE_BASE=https://<域名或IP:端口>`。否則 Bark 通知不帶地圖圖片（詳見下文「架設靜態檔案伺服器」）。
4. 驗證與日常使用同路徑 A（第 2、3 步）。

## 上傳介面

用戶端把擷取的 mysekai 回應主體分片 POST 到 `POST /uploadMySekai`（手動用 curl 依同一協定除錯亦可）。header 如下：

| Header | 說明 |
| --- | --- |
| `X-Upload-Id` | 上傳任務 ID（僅字母數字與 `-` / `_`，長度 1~64），必填 |
| `X-Chunk-Index` | 分片序號，從 0 開始，必填 |
| `X-Total-Chunks` | 總分片數（1~10），必填 |
| `X-Original-Url` | 用戶端原始頁面 URL，用於解析玩家 ID（如 `https://.../user/123456...`）；**選填**，缺失時玩家 ID 記為 `unknown` |
| `X-Script-Version` | 用戶端腳本版本號；伺服器端忽略該 header，可不傳 |

請求主體為原始二進位分片資料（無需 multipart）。

限制：

- 單一檔案總大小 ≤1MB（`MAX_TOTAL_SIZE`）
- 單一分片 ≤1MB（`MAX_CHUNK_SIZE`，超出限制回傳 413）
- 總分片數 ≤10（`MAX_CHUNKS`）

> 注意：總大小上限僅 1MB，**分片大小應明顯小於 1MB 才有意義**（例如 256KB，10 片可傳滿 1MB）。若用戶端用 1MB 分片，任何超過 1MB 的檔案都會從第 2 片起被 413 拒絕，實際上退化成只能單片上傳。

回應：

| 狀態碼 | 含義 |
| --- | --- |
| `200` | 分片已接收，回傳 `OK`；最後一片到達時伺服器端自動完成：合併存檔 → 產生地圖 → 歸檔到 `data/archive/by-id/<user_id>/<時間戳>/` → 推播通知，全程無需人工介入 |
| `400` | 參數不合法（upload id 格式錯誤、分片序號超出範圍、總分片數不在 1~10） |
| `413` | 超過大小限制（單分片超 1MB，或累計總大小超 1MB） |

### curl 範例

存檔 ≤1MB 時單片即可傳完（最常用）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

## Reqable 上報伺服器

不依賴自訂抓封包用戶端，也可以直接用 Reqable 內建的「上報伺服器」功能（Reqable v2.20.0+）：它會把每個已捕獲的 HTTP 工作階段依 [HAR](https://en.wikipedia.org/wiki/HAR_(file_format)) JSON 格式自動 POST 到你的伺服器，可選 gzip / brotli / zstd 壓縮。上報端點**預設開啟**，與分片上傳共存——`python cli.py server` 同時提供兩個介面；設 `REPORT_ENABLED=0` 可關閉：

```bash
python cli.py server
```

設定（`.env`）：

| 變數 | 預設值 | 說明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1`（開啟） | 設 `0` / `false` 關閉上報端點 |
| `REPORT_PATH` | `/reqable/report` | 端點路徑，填入 Reqable 的「上報路徑」 |
| `REPORT_MAX_SIZE` | `8` | HAR 請求主體大小上限（MB）；存檔本身需 ≤1MB，base64 膨脹約 33% |
| `REPORT_TOKEN` | （空） | 選用共享令牌；設定後端點要求請求頭 `X-Report-Token` 匹配 |

每次上報，主機端會：

1. 依 `Content-Encoding`（gzip / br / zstd）解壓並解析 HAR。
2. 遍歷 `log.entries`，取第一個「回應主體（兜底：請求主體）能用 `AES_KEY` / `AES_IV` 解密並解析為 MySekai 存檔」的工作階段——命中規則但與存檔無關的流量會被跳過。
3. 從工作階段 URL 解析玩家 ID（`/user/<id>`，與 `X-Original-Url` 同規則）。
4. 存檔儲存到 `data/raw_mysekai/`，並啟動與分片上傳相同的 生成 → 歸檔 → 推播 流水線。

注意：

- Reqable 每個工作階段**只上報 1 次且失敗不重試**，因此端點會盡快回傳 `200`。請保持服務穩定，並留意 `[REPORT]` 日誌。
- 每次上報只處理 **1 份**存檔（第一個有效條目），因此匹配多個介面的規則不會造成重複推播。
- 安全性：協定本身沒有鑑權。Reqable 無法附加自訂請求頭，建議把隨機字串拼進 `REPORT_PATH`（如 `/reqable/report/9f3a…`），或用反向代理 / 防火牆限制存取，而不是依賴 `REPORT_TOKEN`。

Reqable 側設定範例：

- URL 匹配規則：`https://<遊戲API網域>/*`（或更精確，如 `https://<遊戲API網域>/user/*/mysekai*`）
- 上報路徑：`http://<你的伺服器>:9478/reqable/report`
- 壓縮演算法：gzip / brotli / zstd 皆可（主機端三種都支援）

五個伺服器的遊戲 API 網域：

| 伺服器 | 遊戲 API 網域 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

建議先用網域通配規則 `https://<網域>/*`——無關工作階段會被主機端自動跳過。如果你的伺服器 mysekai 介面也是 `/api/user/*/mysekai*` 路徑（CN 已實測驗證），可收窄為 `https://<網域>/api/user/*/mysekai*` 以減少上報量。

手動 curl 驗證（gzip 壓縮的 HAR）：

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```

## 推播機制

### 預設走 Telegram Bot

- 未在 `config/push_map.json` 中設定的玩家，**一律預設推播到 Telegram**；`push_map.json` 檔案缺失時同樣預設 Telegram。
- Telegram 使用 Bot API `sendMediaGroup`，把 4 張本地 PNG 作為 multipart 直接上傳，**不需要公開網路直連，也不依賴靜態檔案伺服器**；`TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` 缺失時只印出警告並跳過，不影響 Bark 管道。

### Bark 推播依賴公開直連

Bark（Day.app）通知中的圖片是 **URL 直連**：`notify.py` 把圖片位址編碼進 `image=` 參數發給 `api.day.app`，由 Bark 伺服器再去抓取這張圖。因此該 URL 必須**公開網路可達（建議 HTTPS）**，否則 Bark 通知裡沒有圖片。

4 張地圖的直連由 `notify.py` 依以下優先順序拼出：

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| 場景 | base 取值 | 圖片直連形態 |
| --- | --- | --- |
| 伺服器流程（推薦） | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<時間戳>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<時間戳>/site_{5..8}.png` |
| 手動 CLI 推播 | `BARK_IMAGE_BASE` 或 `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`（需把 `data/latest/` 暴露在 `<base>/` 下） |

> 注意：伺服器流程只有在設定了 `BARK_IMAGE_BASE` 時才會拼出帶歸檔路徑的直連；若只設了 `FALLBACK_IMAGE_BASE`，伺服器推播的直連同樣是 `<FALLBACK_IMAGE_BASE>/site_{5..8}.png`。

## 靜態檔案伺服器範例（選填）

目的：把 `data/archive/` 目錄暴露成公開 URL，讓 Bark 伺服器能抓到四張地圖。

**建議做法**：靜態伺服器的根目錄指向專案的 `data/`，再設定 `BARK_IMAGE_BASE=https://<你的域名或IP:端口>`，即可自動映射：

```
data/archive/by-id/<user_id>/<時間戳>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<時間戳>/site_5.png
```

常用範例：

Python 內建（最簡，適合內網/測試）：

```bash
python -m http.server 8000 --directory data
# 然後設定 BARK_IMAGE_BASE=http://<伺服器IP>:8000
```

nginx：

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    # ... ssl 憑證設定 ...
    root /path/to/MySekaiMapper/data;
}
```

Caddy（自動 HTTPS）：

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

注意事項：

- **不要用 `127.0.0.1` / `localhost`** 作為直連位址；Bark 伺服器需要能存取該位址，一般直接選公開網路可達的位址，內網 IP 僅在確認互通時使用。
- **只用 Telegram 則完全不需要靜態伺服器**，跳過本節即可。
- 手動 `cli.py notify` 的直連不帶歸檔路徑，需要另把 `data/latest/` 暴露在 `BARK_IMAGE_BASE` 下；或用 `FALLBACK_IMAGE_BASE` 指向輸出目錄（例如 `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → 該伺服器把 `data/latest/` 掛在 `/output` 下）。

## 玩家推播路由（選填）

在 `config/` 下依需求建立本地設定（格式見同目錄 `*.example.json`，已被 `.gitignore` 忽略）：

- `push_map.json` — 玩家 ID → 推播方式：值為 `"telegram"`、Bark 別名、`"none"`（不推播），也支援組合寫法 `["alias", "telegram"]` 或 `"alias+tg"`。**未設定的玩家預設 `telegram`**。

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark 別名 → 裝置 key：

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 常見問題

- **Bark 通知收不到圖片？** 檢查直連是否公開網路可達：在瀏覽器/手機網路下直接開啟 `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<時間戳>/site_5.png` 應能顯示圖片；內網位址、`127.0.0.1`、或憑證異常的 HTTPS 都會導致抓圖失敗。
- **什麼都沒推播？** 檢查 `push_map.json` 是否把該玩家設成了 `"none"`；只配了 Bark 的使用者是否忘了在該玩家上設定 Bark 別名（未設定的玩家預設走 Telegram）；Telegram 管道是否配了 token 與 chat id；Bark 管道是否缺 key（報 `[BARK] ... failed` 日誌）。
- **不想收到 Bark 只想要 Telegram？** 什麼都不用做——未設定的玩家預設就走 Telegram。

## 命令列工具（cli.py）

所有功能都可透過 `cli.py` 驅動；安裝後（`pip install -e .`）也可用等價的 `mysekai` 命令。命令成功退出碼為 0，出錯為 1（錯誤資訊印到 stderr）。

```bash
python cli.py --help           # 子命令總覽
python cli.py <命令> --help     # 查看某子命令的參數
```

### generate —— 解密存檔並產生地圖

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`：加密存檔路徑（.bin），必填
- 流程：AES-128-CBC 解密 → msgpack 解析 → 擷取掉落座標 → 繪製 4 張地圖（`site_5.png` ~ `site_8.png`）→ 寫出 `rare_resources.txt`
- 輸出到 `data/latest/`，結束時印出實際路徑
- 前置要求：`.env` 已設定 `AES_KEY` / `AES_IV`；存檔中沒有任何掉落點時會報錯退出

### notify —— 推播地圖與統計

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`：包含 `site_*.png` 與 `rare_resources.txt` 的目錄（通常就是 `data/latest/`）
- `[task_id]`：選填，上傳任務 ID，預設 `unknown`。用於從 `data/raw_mysekai/` 反查玩家 ID：優先比對 `mysekai_<玩家ID>_<task_id>.bin`，比對不到時取 raw_mysekai 裡最新的存檔
- 推播到 Telegram 還是 Bark 由 `config/push_map.json` 路由（未設定的玩家預設走 Telegram），詳見「玩家推播路由」

### server —— 啟動上傳服務（分片上傳 + Reqable 上報伺服器）

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- 啟動 FastAPI 服務，用戶端向 `POST /uploadMySekai` 分片上傳加密存檔（介面細節見「上傳介面」）
- 全部片到達後自動完成：合併存檔 → 產生地圖 → 歸檔到 `data/archive/by-id/<user_id>/<時間戳>/` → 依玩家路由推播通知，無需人工介入
- 預設監聽 `9478` 連接埠；公開網路部署時建議透過反向代理暴露為 HTTPS，用戶端腳本中寫死的上傳 URL（含連接埠）需與你的實際部署保持一致

### 典型手動流程

```bash
python cli.py generate mysekai_xxx.bin       # 1. 產生地圖到 data/latest/
python cli.py notify data/latest <task_id>   # 2. 推播（task_id 填上傳 ID，如 chfto53c3）
```

## 目錄結構

```
├── app/                       # 核心套件
│   ├── config.py              # 路徑／環境變數／本地設定集中管理
│   ├── crypto.py              # MySekai 存檔 AES-128-CBC 解密
│   ├── parser.py              # msgpack 解析＋站點座標旋轉（純函式）
│   ├── render.py              # 擷取掉落點 → matplotlib 繪圖＋稀有資源統計
│   ├── notify.py              # 推播：Telegram 媒體群組／Bark，依玩家路由
│   ├── server.py              # FastAPI 分片上傳服務
│   └── cli.py                 # 命令列入口
├── assets/                    # 靜態資源（提交到儲存庫）
│   ├── resourceId.csv         # 物品 ID → 名稱＋圖示（base64）
│   └── NotoSansSC-Regular.ttf # 中文字型（OFL 授權）
├── config/                    # 本地設定（真實檔案不提交，參考 *.example.json）
│   ├── bark_map.example.json  # Bark 別名 → 裝置 key 範本
│   └── push_map.example.json  # 玩家 ID → 推播方式範本
├── data/                      # 執行時期資料（整個目錄 gitignore）
│   ├── tmp/                   # 分片上傳暫存，合併後即清
│   ├── raw_mysekai/           # 合併後的原始（加密）存檔，永久保留
│   ├── archive/               # 歷史成品歸檔 by-id/<user>/<時間戳>/（Bark 直連即指向此處）
│   └── latest/                # 最近一次產生的成品
├── cli.py                     # 統一入口
├── tests/                     # 單元測試（pytest）
├── .env.example               # 環境變數範本（複製為 .env 填寫）
└── requirements.txt           # 執行時期依賴（精確鎖定版本）
```

## 測試

```bash
python -m pytest
```

## 免責聲明

本工具僅用於個人學習與娛樂，請勿用於任何商業用途或違反遊戲服務條款的行為。遊戲資料與美術資源版權歸原版權方所有。

## 授權條款

本專案程式碼採用 [MIT License](LICENSE)（版權所有 © 2025 mouse233），可自由使用、修改與再散布，詳見 [LICENSE](LICENSE)。

> ⚠️ 授權條款僅涵蓋本專案程式碼：`assets/` 中的遊戲素材（如 `resourceId.csv` 內的物品圖示）與遊戲資料版權歸 SEGA / Colorful Palette 等原版權方所有，**不在 MIT 授權範圍內**，請勿將其用於本工具之外的用途。
