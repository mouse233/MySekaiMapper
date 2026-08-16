# 静的ファイルサーバーの例（任意）

目的：`data/archive/` ディレクトリを公開 URL として公開し、Bark サーバーが 4 枚のマップを取得できるようにします。

**推奨方法**：静的サーバーのルートディレクトリをプロジェクトの `data/` に向け、`BARK_IMAGE_BASE=https://<あなたのドメインまたはIP:ポート>` を設定すれば、自動的にマッピングされます：

```
data/archive/by-id/<user_id>/<タイムスタンプ>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<タイムスタンプ>/site_5.png
```

## よく使う例

Python 内蔵（最小構成、内部ネットワーク/テスト向け）：

```bash
python -m http.server 8000 --directory data
# その後 BARK_IMAGE_BASE=http://<サーバーIP>:8000 を設定
```

nginx：

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    # ... ssl 証明書の設定 ...
    root /path/to/MySekaiMapper/data;
}
```

Caddy（自動 HTTPS）：

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

## 注意事項

- 直リンクのアドレスに **`127.0.0.1` / `localhost` は使わないでください**。Bark サーバーがそのアドレスにアクセスできる必要があるため、通常は公開ネットワークから到達可能なアドレスを選びます。内部ネットワークの IP は疎通を確認できた場合のみ使用します。
- **Telegram のみを使う場合は静的サーバーは一切不要**なので、この節は読み飛ばして構いません。
- 手動の `cli.py notify` の直リンクにはアーカイブパスが含まれないため、別途 `data/latest/` を `BARK_IMAGE_BASE` 配下に公開する必要があります。または `FALLBACK_IMAGE_BASE` を出力ディレクトリに向けます（例：`FALLBACK_IMAGE_BASE=http://<host>:5500/output` → そのサーバーが `data/latest/` を `/output` 配下にマウントします）。
