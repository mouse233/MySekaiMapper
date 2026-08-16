---
layout: home

hero:
  name: MySekaiMapper
  text: MySekai 採集點地圖產生工具
  tagline: 《世界計畫 繽紛舞台！feat. 初音未來》（Project Sekai）MySekai 採集點地圖產生與自動推播工具。
  actions:
    - theme: brand
      text: 快速上手
      link: /zh-TW/guide/introduction
    - theme: alt
      text: 在 GitHub 查看
      link: https://github.com/mouse233/MySekaiMapper

features:
  - title: 全自動流水線
    details: 抓封包工具（MitM 模組 / Reqable 上報伺服器）分片上傳 MySekai 資料包，伺服器端自動合併、解密、繪圖並推播，全程無需人工介入。
  - title: 4 張地圖 + 稀有資源統計
    details: 每次任務產生 site_5.png ~ site_8.png（初始空地、心願沙灘、爛漫花田、忘卻之所），外加 rare_resources.txt 稀有資源統計。
  - title: Telegram 優先，Bark 就緒
    details: Telegram 透過 multipart 直傳 4 張 PNG，無需公網直連；設定靜態檔案伺服器後 Bark 即可收到圖片直連通知。
  - title: AES-128-CBC 解密
    details: 解密加密的 MySekai 存檔，解析 msgpack，自動旋轉站點座標，並用 matplotlib 繪製採集地圖。
---

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

> ⚠️ **免責聲明**：本工具僅用於個人學習與娛樂，請勿用於任何商業用途或違反遊戲服務條款的行為。遊戲資料與美術資源版權歸原版權方所有。
