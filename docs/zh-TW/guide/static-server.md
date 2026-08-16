# 靜態檔案伺服器範例（選填）

目的：把 `data/archive/` 目錄暴露成公開 URL，讓 Bark 伺服器能抓到四張地圖。

**建議做法**：靜態伺服器的根目錄指向專案的 `data/`，再設定 `BARK_IMAGE_BASE=https://<你的域名或IP:端口>`，即可自動映射：

```
data/archive/by-id/<user_id>/<時間戳>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<時間戳>/site_5.png
```

## 常用範例

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

## 注意事項

- **不要用 `127.0.0.1` / `localhost`** 作為直連位址；Bark 伺服器需要能存取該位址，一般直接選公開網路可達的位址，內網 IP 僅在確認互通時使用。
- **只用 Telegram 則完全不需要靜態伺服器**，跳過本節即可。
- 手動 `cli.py notify` 的直連不帶歸檔路徑，需要另把 `data/latest/` 暴露在 `BARK_IMAGE_BASE` 下；或用 `FALLBACK_IMAGE_BASE` 指向輸出目錄（例如 `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → 該伺服器把 `data/latest/` 掛在 `/output` 下）。
