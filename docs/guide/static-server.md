# Static File Server Examples (optional)

Purpose: expose the `data/archive/` directory as a public URL so the Bark server can fetch the four maps.

**Recommended setup**: point the static server root at the project's `data/`, then set `BARK_IMAGE_BASE=https://<your-domain-or-ip:port>` for automatic mapping:

```
data/archive/by-id/<user_id>/<timestamp>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<timestamp>/site_5.png
```

## Common examples

Python built-in (simplest; LAN/testing):

```bash
python -m http.server 8000 --directory data
# then set BARK_IMAGE_BASE=http://<server-ip>:8000
```

nginx:

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    # ... ssl certificate config ...
    root /path/to/MySekaiMapper/data;
}
```

Caddy (automatic HTTPS):

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

## Notes

- **Don't use `127.0.0.1` / `localhost`** as the link address; the Bark server must be able to reach it. In general, pick a publicly reachable address; LAN IPs only when connectivity is confirmed.
- **Telegram-only users need no static server at all** — skip this section.
- Manual `cli.py notify` links carry no archive path: expose `data/latest/` under `BARK_IMAGE_BASE` separately, or point `FALLBACK_IMAGE_BASE` at the output directory (e.g. `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → that server mounts `data/latest/` at `/output`).
