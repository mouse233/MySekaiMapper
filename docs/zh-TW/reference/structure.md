# 目錄結構

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
