<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 目錄結構

```text
.
├── cmd/mysekaimapper/       # CLI 進入點
├── internal/
│   ├── har/                 # Reqable HAR 解析與解壓縮
│   ├── mapper/              # AES、MsgPack、資源與渲染
│   ├── notify/              # Telegram 與 Bark 傳送
│   ├── server/              # 上傳與報告 HTTP 端點
│   └── service/             # 佇列、儲存與封存管線
├── assets/                  # 字型與資源圖示
├── config/                  # 本機路由範本
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # 忽略的執行階段資料
│   ├── tmp/                 # 上傳暫存區
│   ├── raw_mysekai/         # 加密來源封存檔
│   ├── archive/             # 依玩家與時間戳記保存的歷史產物
│   └── latest/              # 最新產生的產物
├── docs/                    # VitePress 文件
├── go.mod / go.sum          # Go 模組定義
└── .env.example             # 設定範本
```

`data/`、`.env`、`config/bark_map.json` 與 `config/push_map.json` 是私密的執行階段資料，且會被 Git 忽略。
