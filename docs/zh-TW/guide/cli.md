# 命令列工具（cli.py）

所有功能都可透過 `cli.py` 驅動；安裝後（`pip install -e .`）也可用等價的 `mysekai` 命令。命令成功退出碼為 0，出錯為 1（錯誤資訊印到 stderr）。

```bash
python cli.py --help           # 子命令總覽
python cli.py <命令> --help     # 查看某子命令的參數
```

## generate —— 解密存檔並產生地圖

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`：加密存檔路徑（.bin），必填
- 流程：AES-128-CBC 解密 → msgpack 解析 → 擷取掉落座標 → 繪製 4 張地圖（`site_5.png` ~ `site_8.png`）→ 寫出 `rare_resources.txt`
- 輸出到 `data/latest/`，結束時印出實際路徑
- 前置要求：`.env` 已設定 `AES_KEY` / `AES_IV`；存檔中沒有任何掉落點時會報錯退出

## notify —— 推播地圖與統計

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`：包含 `site_*.png` 與 `rare_resources.txt` 的目錄（通常就是 `data/latest/`）
- `[task_id]`：選填，上傳任務 ID，預設 `unknown`。用於從 `data/raw_mysekai/` 反查玩家 ID：優先比對 `mysekai_<玩家ID>_<task_id>.bin`，比對不到時取 raw_mysekai 裡最新的存檔
- 推播到 Telegram 還是 Bark 由 `config/push_map.json` 路由（未設定的玩家預設走 Telegram），詳見[玩家推播路由](/zh-TW/guide/routing)

## server —— 啟動上傳服務（分片上傳 + Reqable 上報伺服器）

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- 啟動 FastAPI 服務：用戶端向 `POST /uploadMySekai` 上傳加密存檔（單片或分片；介面細節見[上傳介面](/zh-TW/guide/upload-api)）；Reqable 也可把 HAR 工作階段上報到內建上報端點（見[Reqable 上報伺服器](/zh-TW/guide/report-server)）
- 全部片到達後自動完成：合併存檔 → 產生地圖 → 歸檔到 `data/archive/by-id/<user_id>/<時間戳>/` → 依玩家路由推播通知，無需人工介入
- 預設監聽 `9478` 連接埠；公開網路部署時建議透過反向代理暴露為 HTTPS，用戶端腳本中寫死的上傳 URL（含連接埠）需與你的實際部署保持一致

## 典型手動流程

```bash
python cli.py generate mysekai_xxx.bin       # 1. 產生地圖到 data/latest/
python cli.py notify data/latest <task_id>   # 2. 推播（task_id 填上傳 ID，如 chfto53c3）
```
