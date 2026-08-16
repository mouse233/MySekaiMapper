# 静态文件服务器示例（可选）

目的：把 `data/archive/` 目录暴露成公开 URL，让 Bark 服务器能抓到四张地图。

**推荐做法**：静态服务器的根目录指向项目的 `data/`，再设置 `BARK_IMAGE_BASE=https://<你的域名或IP:端口>`，即可自动映射：

```
data/archive/by-id/<user_id>/<时间戳>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<时间戳>/site_5.png
```

## 常用示例

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

## 注意事项

- **不要用 `127.0.0.1` / `localhost`** 作为直链地址；Bark 服务器需要能访问该地址，一般直接选公网可达的地址，内网 IP 仅在确认互通时使用。
- **只用 Telegram 则完全不需要静态服务器**，跳过本节即可。
- 手动 `cli.py notify` 的直链不带归档路径，需要另把 `data/latest/` 暴露在 `BARK_IMAGE_BASE` 下；或用 `FALLBACK_IMAGE_BASE` 指向输出目录（例如 `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → 该服务器把 `data/latest/` 挂在 `/output` 下）。
