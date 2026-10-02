<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# 通知と静的ファイル

`config/push_map.example.json`、`config/bark_map.example.json` からローカル設定を作成します。これらのファイルにはプレイヤー／デバイス識別子が含まれており、Git の追跡対象外です。

### プレイヤールーティング

`config/push_map.json` は、プレイヤー ID を `telegram`、`bark:<エイリアス>`、`astrbot:<エイリアス>`、`none`、接頭辞付きの `+tg` 文字列、またはメソッド配列にマッピングします：

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

利用可能なルーティング値がないプレイヤーは、デフォルトで Telegram に送信されます。

Bark のルートには `bark:<エイリアス>` が必須です。AstrBot は `astrbot:<エイリアス>` を使います。更新時に `push_map.json` の `"sls"`、`"xufan"` を `"bark:sls"`、`"bark:xufan"` に変更してください。`bark_map.json` のキーは `"sls"`、`"xufan"` のままです。接頭辞のないルートはエラーになりますが、他の有効な通知先への送信は続行します。`"bark:klee+tg"` も使えますが、配列 `["bark:klee", "telegram"]` を推奨します。

### Telegram

Telegram は、生成された通常の `site_*.png` をすべて、ローカルの multipart メディアグループとしてアップロードします。`TELEGRAM_BOT_TOKEN` と `TELEGRAM_CHAT_ID` が必要ですが、公開画像サーバーは不要です。Telegram が失敗しても、設定済みの Bark への送信は妨げられません。

### AstrBot Push Lite（QQ）

既存の AstrBot に [astrbot_plugin_push_lite](https://github.com/Raven95676/astrbot_plugin_push_lite) をインストールします。QQ は OneBot v11 で接続し、プラグインの API token とポート（既定 `9966`）を設定してください。`config/astrbot_map.example.json` から `config/astrbot_map.json` を作成します：

```json
{
  "qq_me": {"platform_id": "qq_main", "type": "private", "qq": "123456789"},
  "qq_group": {"platform_id": "qq_main", "type": "group", "qq": "987654321"},
  "session": {"umo": "qq_main:FriendMessage:123456789"}
}
```

`platform_id` は AstrBot の実際のボット／プラットフォームインスタンス ID です。`aiocqhttp` やボットの QQ 番号とは限りません。`type` は `private` または `group`、`qq` は宛先の QQ 番号またはグループ番号を表す**文字列**です。`platform_id:FriendMessage:qq` または `platform_id:GroupMessage:qq` が自動生成されます。`/sid` でプラットフォーム ID を一度確認すれば、各受信者がコマンドを実行する必要はありません。完全な `umo` だけを指定する方法も使えますが、他の3項目とは併用できません。[AstrBot のセッション仕様](https://docs.astrbot.app/use/command.html#name)も参照してください。

`config/push_map.json` に `astrbot:<エイリアス>` を指定します。既存の通知先との併用も可能です：

```json
{
  "1234567890123456789": ["astrbot:qq_me", "astrbot:qq_group"],
  "1234567890123456790": ["astrbot:qq_me", "telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

`.env` に以下を設定します：

```dotenv
ASTRBOT_PUSH_URL=http://astrbot:9966
ASTRBOT_PUSH_TOKEN=your-plugin-api-token
ASTRBOT_ALLOW_INSECURE_HTTP=1
```

`ASTRBOT_PUSH_URL` は `/send` を含まないサービスの**ベース URL**です。リバースプロキシのパス接頭辞に対応します。既定では HTTPS が必要です。信頼できるローカル／Docker ネットワークでのみ `ASTRBOT_ALLOW_INSECURE_HTTP=1` を使ってください。別コンテナからは共有ネットワーク上の AstrBot サービス名を指定します。`localhost` は現在のコンテナです。ホスト上のサービスには到達可能なアドレスとプラグインの公開ポートを使います。Bearer token を送信し、リダイレクトには従いません。

`serve` と `notify` は各宛先にテキスト概要を送り、続いて通常の `site_*.png` を Base64 画像として1枚ずつ送信します。公開画像サーバーや共有ディレクトリは不要です。画像はエンコード前に1枚8 MiBまでです。プロキシの要求サイズ制限も調整してください。同じエイリアスの重複は除去され、失敗しても他の宛先や通知経路への送信を続けます。

成功応答は **Push Lite のキューへの登録**を表し、QQ への配信確認ではありません。ログには `request queued` と記録します。配信コールバックや自動再試行は使用しません。キューはメモリ内にあるため、再起動で未送信メッセージが失われる場合があります。配信結果は AstrBot ログで確認してください。ルート未設定のプレイヤーは引き続き Telegram が既定です。`.env` と `config/astrbot_map.json` は Git の対象外です。

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
