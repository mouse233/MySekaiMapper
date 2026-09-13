<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# MySekaiMapper

🌐 **Languages**: [English](../../) · [简体中文](../../zh-CN/) · [繁體中文](../../zh-TW/) · [日本語](../../ja-JP/) · [한국어](../../ko-KR/)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/zh-TW/>

這是一項 Go 服務，可將已加密的 *Project SEKAI* MySekai 存檔轉換為資源採集地圖，並將結果傳送至 Telegram 或 Bark（Day.app）。

它可搭配 MitM 擷取用戶端或 Reqable 的 **Report Server** 使用：擷取工具上傳 MySekai 存檔後，服務會解密並解析內容、繪製地圖與稀有資源摘要、封存產物，並自動發送通知，無須手動處理。

一般的 MySekai 區域會產生 `site_5.png`（草原）、`site_6.png`（海灘）、`site_7.png`（花園）、`site_8.png`（紀念場所）及 `rare_resources.txt`。渲染器與通知程式亦能處理其他一般的 `site_*.png` 輸出。

此擷取流程已在 Nuverse 營運的 CN 與 TW 伺服器上驗證。其他地區是否可用，取決於其 API 路徑與存檔格式。

## 運作方式

```text
Game API response → MitM module / Reqable Report Server
    │  ① POST /uploadMySekai (single upload or ordered chunks)
    │  ② POST /reqable/report (HAR, optionally gzip / br / zstd)
    ▼
mysekaimapper serve
    ├─ AES-128-CBC decrypt + MsgPack parse + coordinate normalization
    ├─ render site_*.png + rare_resources.txt
    ├─ archive data/archive/by-id/<player_id>/<timestamp>/
    └─ publish data/latest/ and notify
         ├─ Telegram: upload local images as multipart media groups
         └─ Bark: send image URLs from a public static-file server
```
