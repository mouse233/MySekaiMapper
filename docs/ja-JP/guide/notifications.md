<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# 通知と静的ファイル

`config/push_map.example.json`、`config/bark_map.example.json` からローカル設定を作成します。これらのファイルにはプレイヤー／デバイス識別子が含まれており、Git の追跡対象外です。

### プレイヤールーティング

`config/push_map.json` は、プレイヤー ID を `telegram`、Bark エイリアス、`none`、`+tg` 文字列、またはメソッド配列にマッピングします：

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "klee"],
  "1234567890123456791": "none"
}
```

利用可能なルーティング値がないプレイヤーは、デフォルトで Telegram に送信されます。

### Telegram

Telegram は、生成された通常の `site_*.png` をすべて、ローカルの multipart メディアグループとしてアップロードします。`TELEGRAM_BOT_TOKEN` と `TELEGRAM_CHAT_ID` が必要ですが、公開画像サーバーは不要です。Telegram が失敗しても、設定済みの Bark への送信は妨げられません。

### Bark

Bark はレアリソースの概要を送信し、生成された通常の `site_*.png` についてもそれぞれ通知します。`config/bark_map.json` でエイリアスとデバイスキーをマッピングします：

```json
{ "klee": "paste-your-bark-key-here" }
```

Bark は画像 URL を自動的に取得します。自動サービスのタスクでは、`BARK_IMAGE_BASE` を公開された `data/` ルートに設定します。アーカイブ URL は次のとおりです：

```text
https://maps.example.com/archive/by-id/<player_id>/<timestamp>/site_5.png
```

手動の `notify` における画像ルートの優先順位は、`--image-base`、`BARK_IMAGE_BASE`、`FALLBACK_IMAGE_BASE` の順です。このルートでは、選択した出力ディレクトリを直接公開する必要があります。

### 静的ファイルサーバー

Bark の画像には `localhost` や `127.0.0.1` は使用できません。次のような公開 HTTPS を使用してください：

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    root /path/to/MySekaiMapper/data;
}
```

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

通知器は出力ディレクトリ内のシンボリックリンクを無視し、認証情報や完全な通知 URL を記録しません。
