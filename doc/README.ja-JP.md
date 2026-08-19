# MySekaiMapper

🌐 **Languages**: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/ja-JP/>

暗号化された *Project SEKAI* の MySekai セーブデータを資源収集マップへ変換し、結果を Telegram または Bark（Day.app）へ送信する Go サービスです。

MitM キャプチャクライアントまたは Reqable の **Report Server** と連携します。キャプチャツールが MySekai セーブデータをアップロードすると、サービスが復号・解析してマップとレアリソース概要を描画し、成果物をアーカイブして、手動処理なしで通知を配信します。

通常の MySekai エリアでは `site_5.png`（草原）、`site_6.png`（浜辺）、`site_7.png`（花畑）、`site_8.png`（記念所）、および `rare_resources.txt` が生成されます。レンダラーと通知機能は、追加の通常 `site_*.png` 出力にも対応しています。

キャプチャフローは Nuverse が運営する中国（CN）および台湾（TW）サーバーで検証されています。他リージョンで利用できるかどうかは、その API パスとセーブデータ形式に依存します。

## 仕組み

```text
Game API response → MitM module / Reqable Report Server
    │  ① POST /uploadMySekai (single upload or ordered chunks)
    │  ② POST /reqable/report (HAR, optionally gzip / br / zstd)
    ▼
mysekaimapper serve
    ├─ AES-128-CBC decrypt + MsgPack parse + coordinate normalization
    ├─ render site_*.png + rare_resources.txt
    ├─ archive data/archive/by-id/<player_id>/<timestamp>/
    └─ publish data/latest/ and notify
         ├─ Telegram: upload local images as multipart media groups
         └─ Bark: send image URLs from a public static-file server
```

## クイックスタート

構成に合う通知経路を選択してください。

- **経路 A — Telegram のみ**：最も簡単な選択肢です。プレイヤーのルーティングファイルや公開画像サーバーは必要ありません。
- **経路 B — Bark を有効化**：Bark キー、プレイヤールーティング、画像用の公開静的ファイルサーバーを設定します。

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
| `BARK_ICON` | 任意 | Bark 通知に含めるアイコン URL |
| `BARK_IMAGE_BASE` | Bark 画像 | アーカイブしたマップ画像の公開ベース URL |
| `FALLBACK_IMAGE_BASE` | 任意 | `BARK_IMAGE_BASE` 未設定時の画像ベース URL |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | 任意 | Reqable レポートエンドポイントの設定 |
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
2. `config/push_map.example.json` から `config/push_map.json` を作成し、プレイヤー ID を Bark エイリアス、`telegram`、`none`、またはそれらの組み合わせへ対応付けます。

    ```json
    {
      "1234567890123456789": ["klee"],
      "1234567890123456790": ["telegram", "klee"],
      "1234567890123456791": "none"
    }
    ```

3. リポジトリの `data/` ディレクトリを公開 HTTP(S) 静的ファイルサーバーで配信し、その公開ルートを `BARK_IMAGE_BASE` に設定します。

    ```dotenv
    BARK_IMAGE_BASE=https://maps.example.com
    ```

未設定のプレイヤーは Telegram が既定です。そのため Telegram が未設定の場合、未設定プレイヤーには通知されません。Bark 専用で使う場合は、Bark エイリアスを明示的に割り当ててください。

## サービスの実行

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

サーバーは準備完了 URL を出力し、アップロード／レポートの受け付け、キュー投入、解析、描画、アーカイブ、通知、経過時間、タスク ID、`player_id` のライフサイクルログを書き出します。アーカイブ本文、シークレット、トークン、完全な通知 URL は意図的にログへ記録しません。

プロセスは `SIGINT` と `SIGTERM` を処理します。HTTP リクエストの受け付けを停止した後、すでに受け付けたジョブを最大 15 秒間処理します。

コンパイル済みバイナリは `--root /path/to/MySekaiMapper` を指定すればチェックアウト外から実行できます。指定しない場合は、作業ディレクトリからリポジトリルートを検出します。

## Upload API

`POST /uploadMySekai` は暗号化された MySekai レスポンス本文を直接受け取ります。通常は単一アップロードで十分ですが、キャプチャクライアントとの互換性のため、順序付きチャンクも引き続きサポートされています。

| ヘッダー | 必須 | 説明 |
| --- | --- | --- |
| `X-Upload-Id` | はい | `^[A-Za-z0-9_-]{1,64}$` に一致するタスク識別子 |
| `X-Chunk-Index` | はい | 0 始まりのチャンク番号 |
| `X-Total-Chunks` | はい | 1～10 の総チャンク数 |
| `X-Original-Url` | いいえ | 元のゲーム URL。`/user/<id>` がプレイヤールートを提供します |
| `X-Script-Version` | いいえ | キャプチャクライアントとの互換性のため受け付け、サービスでは無視します |

暗号化アーカイブ、各チャンク、および結合済みアップロードはいずれも 1 MiB に制限されます。リクエストが正常に受理されるとプレーンテキストの `OK` が返され、描画と通知はバックグラウンドで続行されます。

### 単一アップロードの例

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### チャンクアップロードの例

共通の `X-Upload-Id`、順序どおりの番号、最大 10 個のチャンクを使用します。

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_
total=$(ls /tmp/ms_chunk_* | wc -l | tr -d ' ')

i=0
for chunk in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
    --data-binary @"$chunk"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

一般的なレスポンスは、受理済みアップロードの `200 OK`、無効な識別子またはチャンク範囲の `400 Bad Request`、サイズ制限超過の `413 Payload Too Large`、必須アップロードヘッダーの欠落または非整数値の `422 Unprocessable Entity` です。

## Reqable Report Server

Reqable v2.20.0以降では、キャプチャした各HTTPセッションをHAR JSONとしてこのサービスにPOSTできます。レポートエンドポイントはデフォルトで有効になっており、`/uploadMySekai`と併用できます。

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| 変数 | デフォルト値 | 説明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | レポートを無効にするには `0`、`false`、`no`、または `off` を設定 |
| `REPORT_PATH` | `/reqable/report` | Reqableで設定するエンドポイントパス |
| `REPORT_MAX_SIZE` | `1` | 展開後のHAR本文の最大サイズ（MiB） |
| `REPORT_TOKEN` | 空 | `X-Report-Token` に必要なオプション値 |

### 処理フロー

各レポートについて、サービスは次の処理を行います。

1. `identity`、`gzip`、`br`、`zstd`、または `zstandard` のコンテンツを展開し、HARを解析します。content-sizeフィールドを持たないストリーミングzstdフレームにも対応しています。
2. `log.entries` を走査し、`AES_KEY`/`AES_IV` で復号でき、かつMySekaiアーカイブとして検証に成功した最初のレスポンス本文（該当しない場合はリクエスト本文）を受け付けます。
3. 一致したセッションURLの `/user/<id>` から `player_id` を抽出します。
4. 暗号化されたアーカイブを `data/raw_mysekai/` に保存し、アップロードで使用されるものと同じ render → archive → notify パイプラインを開始します。

> Reqableは各セッションを1回だけレポートし、再試行しません。サービスを稼働状態に保ち、`[REPORT]` ログを監視してください。構文的に有効なHARであれば、MySekaiアーカイブが含まれていなくても `ok` が返されます。レポート内で最初に見つかった有効なアーカイブのみが処理されます。

### Reqableの設定

- **マッチングルール**：`https://<game-api-domain>/api/user/*/mysekai*`
- **サーバー URL**：`http://<your-server>:9478/reqable/report`（またはカスタムの `REPORT_PATH`）

| サーバー | ゲームAPIドメイン |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

このマッチングパターンはCNで検証済みです。お使いの地域で別のMySekai APIパスが使用されている場合は、キャプチャしたURLを確認し、ルールを調整してください。

### セキュリティ

Reqableではカスタムの `X-Report-Token` ヘッダーを追加できません。`/reqable/report/<random>` のような長くランダムな `REPORT_PATH` を使用し、リバースプロキシまたはファイアウォールでアクセスを制限してください。適切な制御なしにデフォルトのエンドポイントを公開しないでください。

### gzip HARの手動テスト

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```

## 通知と静的ファイル

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

## コマンドラインリファレンス

まずバイナリをビルドします。

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

すべてのコマンドはデフォルトで `.env` を読み込み、`--env /path/to/file` を受け付けます。`--root` はサブコマンドの後であればどこにでも指定できます。

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

セーブデータを復号・解析し、マップを書き出さずに安全な集計 JSON 概要を表示します。

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

アーカイブを復号してドロップを抽出し、`site_*.png` と `rare_resources.txt` を書き出します。`--output` の既定値は `data/latest` です。`--assets` でアセットディレクトリを上書きできます。

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

`--output` は必須です。`--task-id` と `--player-id` の既定値は `unknown` です。プレイヤー固有のルーティングが必要な場合は、必ず実際のプレイヤー ID を渡してください。

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

アップロードおよびレポート用 HTTP エンドポイントを起動します。既定値は `0.0.0.0:9478` です。

## ディレクトリ構成

```text
.
├── cmd/mysekaimapper/       # CLI entry point
├── internal/
│   ├── har/                 # Reqable HAR parsing and decompression
│   ├── mapper/              # AES, MsgPack, resources, and rendering
│   ├── notify/              # Telegram and Bark delivery
│   ├── server/              # Upload and report HTTP endpoints
│   └── service/             # Queue, storage, and archive pipeline
├── assets/                  # Font and resource icons
├── config/                  # Local routing templates
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # Ignored runtime data
│   ├── tmp/                 # Upload staging
│   ├── raw_mysekai/         # Encrypted source archives
│   ├── archive/             # Historical artifacts by player and timestamp
│   └── latest/              # Latest generated artifacts
├── docs/                    # VitePress documentation
├── go.mod / go.sum          # Go module definition
└── .env.example             # Configuration template
```

`data/`、`.env`、`config/bark_map.json`、`config/push_map.json` は非公開の実行時データであり、Git では無視されます。

## テスト

```bash
go test ./...
go build -o /tmp/mysekaimapper ./cmd/mysekaimapper
npm run docs:build
```

GitHub Actions は、push とプルリクエストに対して Go のテストスイートおよびビルドを実行します。

## Go リファクタリング

現在のランタイムは Go のみで構成されています。モジュールは `cmd/`、`internal/`、`go.mod`、`go.sum` による標準のルートレイアウトに従っています。Python のソース、依存関係、CI は削除されました。アーカイブ済みの参照実装は、[`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) ブランチおよび [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) タグに残されています。

HTTP エンドポイント、環境変数、出力名、アーカイブレイアウト、ルーティングファイル形式は互換性を維持しています。Go レンダラーは固定キャンバスを使用するため、生成される PNG が以前の Matplotlib 出力とピクセル単位で完全に同一になる保証はありません。

## 免責事項

このツールは個人的な学習および娯楽目的に限って使用してください。商用目的、またはゲームの利用規約に違反する方法で使用しないでください。ゲームデータおよびアセットは、それぞれの権利者に帰属します。

## ライセンス

プロジェクトのコードは [MIT](https://github.com/mouse233/MySekaiMapper/blob/feat/go-rewrite/LICENSE) ライセンス（Copyright © 2025 mouse233）で提供されます。`assets/` 配下のゲームアセットおよびゲームデータはそれぞれの権利者に帰属し、本ライセンスの対象外です。
