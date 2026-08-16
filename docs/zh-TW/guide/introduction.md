# 專案介紹

**MySekaiMapper** 是《世界計畫 繽紛舞台！feat. 初音未來》（Project Sekai）MySekai（我的世界）採集點地圖產生工具。

**專案初衷**：搭配 MitM 模組或 Reqable 的「上報伺服器」功能使用——抓封包工具擷取遊戲內 MySekai 資料封包後，自動分片上傳到本服務；伺服器端合併加密存檔、解密並擷取各站點的資源掉落座標，繪製採集地圖，再把結果（含稀有資源統計）推播到玩家的 Telegram / Bark（iOS Day.app），全程無需人工介入。

一次任務會產生 **4 張地圖**：`site_5.png`（初始空地）、`site_6.png`（心願沙灘）、`site_7.png`（爛漫花田）、`site_8.png`（忘卻之所），外加一份 `rare_resources.txt` 稀有資源統計。

::: info 伺服器相容性
本專案已在朝夕光年（Nuverse）營運的 CN 服 / TW 服中測試通過，其他伺服器可用性未知。
:::

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
             ├─ Telegram  ：圖片 multipart 直傳，無需公網直連 ← 預設管道
             └─ Bark      ：以 image= URL 直連通知，需靜態檔案伺服器
```

## 環境需求

- Python 3.10+
- 執行時期依賴以 `requirements.txt` 為準（精確鎖定版本）

## 快速導覽

- [快速上手](/zh-TW/guide/quickstart) — 安裝、設定 `.env`、選擇路徑 A 或路徑 B
- [上傳介面](/zh-TW/guide/upload-api) — 供抓封包用戶端使用的分片上傳介面
- [推播機制](/zh-TW/guide/push) — Telegram / Bark 通知如何運作
- [命令列工具](/zh-TW/guide/cli) — `cli.py generate` / `notify` / `server`
