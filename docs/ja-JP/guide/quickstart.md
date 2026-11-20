<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# クイックスタート

構成に合う通知経路を選択してください。

- **経路 A — Telegram のみ**：最も簡単な選択肢です。プレイヤーのルーティングファイルや公開画像サーバーは必要ありません。
- **経路 B — Bark を有効化**：Bark キー、プレイヤールーティング、画像用の公開静的ファイルサーバーを設定します。
- **経路 C — AstrBot（QQ）**：下の手順 5 で Push Lite と送信先エイリアスを設定します。公開画像サーバーは不要です。

### 1. 必要要件とビルド

Go **1.25 以降**が必要です。

```bash
go version
cp .env.example .env
go test ./...
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

`.env` の `AES_KEY` と `AES_IV` には、16 バイトの AES-128-CBC 値が必要です。`.env` やローカルのルーティングファイルをコミットしないでください。

### 2. `.env` を設定する

| 変数 | 必須 | 説明 |
| --- | --- | --- |
| `AES_KEY`, `AES_IV` | はい | 16 バイトの MySekai AES-128-CBC キーおよび IV |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` | Telegram のみ | [@BotFather](https://t.me/BotFather) から取得する Bot 認証情報と送信先チャット ID |
| `ASTRBOT_PUSH_URL`, `ASTRBOT_PUSH_TOKEN` | AstrBot のみ | Push Lite のベース URL と API token。通知設定を参照 |
| `BARK_ICON` | 任意 | Bark 通知に含めるアイコン URL |
| `BARK_IMAGE_BASE` | Bark 画像 | アーカイブしたマップ画像の公開ベース URL |
| `FALLBACK_IMAGE_BASE` | 任意 | `BARK_IMAGE_BASE` 未設定時の画像ベース URL |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | 任意 | Reqable レポートエンドポイントの設定 |
| `UPLOAD_TOKEN` | 任意 | 設定すると、`/uploadMySekai` は一致する `X-Upload-Token` ヘッダーを要求します |
| `MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, `MYSK_DATA_DIR` | 任意 | デフォルトのリポジトリディレクトリを上書き |

### 3. 経路 A — Telegram のみ

1. `.env` に Telegram の変数を設定します。

    ```dotenv
    TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
    TELEGRAM_CHAT_ID=123456789
    ```

2. 既存の暗号化済みセーブデータで、解析と通知を必要に応じて確認します。

    ```bash
    bin/mysekaimapper generate --input data/raw_mysekai/mysekai.bin
    bin/mysekaimapper notify \
      --output data/latest \
      --task-id manual-001 \
      --player-id 1234567890123456789
    ```

3. 通常運用のためにサービスを起動します。

    ```bash
    bin/mysekaimapper serve --host 0.0.0.0 --port 9478
    ```

`config/push_map.json` に存在しないプレイヤーは、デフォルトで Telegram に送信されます。経路 A では Bark マップ、push マップ、公開画像サーバーのいずれも必要ありません。

### 4. 経路 B — Bark を有効にする

経路 A の設定に加えて（Bark 専用のルートでは Telegram を省略できます）、次を行います。

1. `config/bark_map.example.json` から `config/bark_map.json` を作成し、Bark エイリアスと各デバイスキーを対応付けます。
2. `config/push_map.example.json` から `config/push_map.json` を作成し、プレイヤー ID を `bark:<エイリアス>`、`telegram`、`none`、またはそれらの組み合わせへ対応付けます。

    ```json
    {
      "1234567890123456789": ["bark:klee"],
      "1234567890123456790": ["telegram", "bark:klee"],
      "1234567890123456791": "none"
    }
    ```

3. リポジトリの `data/` ディレクトリを公開 HTTP(S) 静的ファイルサーバーで配信し、その公開ルートを `BARK_IMAGE_BASE` に設定します。

    ```dotenv
    BARK_IMAGE_BASE=https://maps.example.com
    ```

未設定のプレイヤーは Telegram が既定です。そのため Telegram が未設定の場合、未設定プレイヤーには通知されません。Bark 専用で使う場合は、Bark エイリアスを明示的に割り当ててください。

### 5. 経路 C — AstrBot（QQ）

手順 1–2 の完了後、AstrBot を直接設定します。Telegram と Bark は任意です。

1. AstrBot に `astrbot_plugin_push_lite` をインストールし、OneBot v11 で QQ を接続して、プラグインの API token と待ち受けポート（既定 `9966`）を設定します。
2. `.env` に送信サービスを設定します（HTTP は信頼できるローカル環境または Docker のプライベートネットワーク専用です）。

   ```dotenv
   ASTRBOT_PUSH_URL=http://astrbot:9966
   ASTRBOT_PUSH_TOKEN=your-plugin-api-token
   ASTRBOT_ALLOW_INSECURE_HTTP=1
   ```

   MySekaiMapper から接続できるアドレスを使用し、末尾に `/send` を付けないでください。HTTPS では `ASTRBOT_ALLOW_INSECURE_HTTP=0` のままにします。

3. `config/astrbot_map.example.json` から `config/astrbot_map.json` を作成し、送信先を設定します。

   ```json
   {
     "qq_me": {"platform_id": "qq_main", "type": "private", "qq": "123456789"}
   }
   ```

   `qq_main` を AstrBot の実際のプラットフォームインスタンス ID（`/sid` で確認可能）に置き換え、`qq` に送信先の QQ 番号を文字列で指定します。グループの場合は `type` を `group` に変更し、グループ番号を指定します。

4. `config/push_map.json` を作成または更新し、実際のゲームプレイヤー ID を送信先エイリアスに対応付けます。

   ```json
   {
     "1234567890123456789": ["astrbot:qq_me"]
   }
   ```

5. サービスを起動します。

   ```bash
   bin/mysekaimapper serve --host 0.0.0.0 --port 9478
   ```

公開画像サーバーや共有ディレクトリは不要です。ルート未設定のプレイヤーは引き続き Telegram が既定です。Push Lite の成功応答はキューへの登録を示します。QQ への配信結果は AstrBot のログで確認してください。詳細は後述の「通知と静的ファイル」の AstrBot 節を参照してください。
