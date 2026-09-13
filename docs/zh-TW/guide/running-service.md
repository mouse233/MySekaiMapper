<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 執行服務

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

伺服器會印出就緒 URL，並為上傳／報告接受、佇列處理、解析、渲染、封存、通知、耗時、工作 ID 與 `player_id` 寫入生命週期日誌。它刻意不會記錄封存內容、密鑰、權杖或完整的通知 URL。

程序會處理 `SIGINT` 與 `SIGTERM`：先停止接受 HTTP 請求，接著最多等待 15 秒，以排空已接受的工作。

已編譯的二進位檔可在專案工作區外透過 `--root /path/to/MySekaiMapper` 執行；否則會從工作目錄尋找儲存庫根目錄。
