# MySekaiMapper

🌐 言語: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/ja-JP/>

「プロジェクトセカイ カラフルステージ！feat. 初音ミク」（Project Sekai）の MySekai（マイセカイ）採集ポイントマップ生成ツールです。

**開発のきっかけ**：MitM モジュールまたは Reqable の「レポートサーバー」機能と組み合わせて使用します——キャプチャツールがゲーム内の MySekai データパケットを捕捉した後、自動的に本サービスへアップロードします（1 回の POST で送信可能、分割アップロードにも対応）。サーバー側は暗号化セーブデータを復号し、各ステーションの資源ドロップ座標を抽出して採集マップを描画し、その結果（レア資源統計を含む）をプレイヤーの Telegram / Bark（iOS Day.app）へプッシュします。一切の手動介入は不要です。

1 回のタスクで **4 枚のマップ**が生成されます：`site_5.png`（さいしょの原っぱ）、`site_6.png`（願いの砂浜）、`site_7.png`（彩りの花畑）、`site_8.png`（忘れ去られた場所）。さらに `rare_resources.txt` のレア資源統計も出力します。

本プロジェクトは朝夕光年（Nuverse）が運営する CN サーバー / TW サーバーで動作確認済みです。他のサーバーでの動作は未確認です。

## ワークフロー

```
ゲーム API 応答 → MitM モジュール / Reqable レポートサーバー（mysekai データをキャプチャ）
   │  ① 自動アップロード（1 回の POST、分割にも対応）→ server.py が自動処理
   │  ② または .bin セーブデータを手動配置 → cli.py generate
   ▼
parser.py    AES-128-CBC 復号 + msgpack 解析 + 座標回転
   ▼
render.py    site_5.png ~ site_8.png + rare_resources.txt を描画 → data/latest/
   ▼
notify.py    プッシュ：
             ├─ Telegram  ：画像を multipart で直接送信、公開直リンク不要 ← デフォルトチャネル
             └─ Bark      ：image= URL 直リンクで通知、静的ファイルサーバーが必要
```

## クイックスタート

まずインストールと `.env` の基本設定を済ませてから、希望するプッシュ方法に応じてパスを選択してください：

- **パス A（Telegram Bot プッシュのみ）**：設定が最も少なく、まずこちらで一通り動かすことをお勧めします。
- **パス B（Bark プッシュを有効化）**：パス A に加えて、Bark key、プレイヤールーティング、静的ファイルサーバーの追加設定が必要です。

### 1. インストール

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 任意:mysekai コマンドをインストール(python cli.py ... と同等)
venv/bin/pip install -e .
```

### 2. .env の設定（必須項目）

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV` は MySekai セーブデータの AES-128-CBC 復号キー（各 16 バイト）で、どのパスを選んでも必ず設定する必要があります。その他の変数は選択したパスに応じて設定します：

| 変数 | 必須 | 説明 |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | MySekai セーブデータの AES-128-CBC キー、各 16 バイト |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | 任意* | Telegram プッシュ（デフォルトチャネル）に必要。[@BotFather](https://t.me/BotFather) から取得 |
| `BARK_ICON` | 任意 | Bark 通知アイコンの URL |
| `BARK_IMAGE_BASE` | 任意 | 静的ファイルサーバーのルートアドレス（Bark の画像直リンク送信用、後述） |
| `FALLBACK_IMAGE_BASE` | 任意 | `BARK_IMAGE_BASE` 未設定時の画像直リンクのフォールバックアドレス |

> \* Bark だけで通知を受け取りたい場合：Telegram の設定は空のままで構いませんが、**`config/push_map.json` でプレイヤーを Bark エイリアスにルーティングする必要があります**。そうしないと、未設定のプレイヤーはデフォルトで Telegram へ送られますが、Telegram の設定が不足している場合は警告が 1 行出力されてスキップされ、結果として何もプッシュされません。

### 3. パス A：Telegram Bot プッシュのみ（最小構成）

適用シーン：Telegram でマップと統計を受け取れればよく、他のコンポーネントには手を出したくない場合。

1. `.env` に Telegram の設定を記入します（[@BotFather](https://t.me/BotFather) から取得）：

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 手動で一度実行して動作を確認します：

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. 日常利用：アップロードサーバーを起動すると、セーブデータの到着後に自動でマップが生成されプッシュされます。キャプチャ方法は次の 2 通りです：

   - **MitM モジュール**：「アップロードAPI」に従ってアップロード
   - **Reqable レポートサーバー**：マッチングルールとアップロードパスを設定（後述の「Reqable レポートサーバー」の章を参照）

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

パス A で**不要なもの**：`config/push_map.json`、`config/bark_map.json`、静的ファイルサーバー、`BARK_IMAGE_BASE`。未設定のプレイヤーはデフォルトで Telegram にプッシュされます。

### 4. パス B：Bark プッシュを有効化（追加設定が必要）

パス A をベースに（Telegram の設定はそのままでも、空のままにして Bark のみでも可）、以下の順に設定を追加します：

1. **Bark key の設定**：`config/bark_map.json` で各エイリアスにデバイス key を設定します（テンプレートは同ディレクトリの `bark_map.example.json` を参照）。
2. **プレイヤールーティングの設定**：`config/push_map.json` でプレイヤー ID を Bark エイリアスにルーティングします。例：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ⚠️ **必須設定**：未設定のプレイヤーはデフォルトで Telegram に送られます。その時点で Telegram も未設定の場合は、警告が出力されてスキップされ、結果として何もプッシュされません。
3. **静的ファイルサーバーの構築**：プロジェクトの `data/` ディレクトリを公開ネットワークからアクセス可能な HTTP(S) サービスとして公開し、`.env` に `BARK_IMAGE_BASE=https://<ドメインまたはIP:ポート>` を設定します。そうしないと Bark 通知にマップ画像が含まれません（詳細は後述の「静的ファイルサーバーの例」を参照）。
4. 検証と日常利用はパス A（手順 2、3）と同じです。

## アップロードAPI

クライアントはキャプチャした mysekai レスポンスボディを `POST /uploadMySekai` にアップロードします（1 回の POST で送信可能。分割アップロードは互換性のため残しています）。手動で curl を使い同じプロトコルでデバッグすることも可能です。ヘッダーは以下の通りです：

| Header | 説明 |
| --- | --- |
| `X-Upload-Id` | アップロードタスク ID（英数字と `-` / `_` のみ、長さ 1〜64）、必須 |
| `X-Chunk-Index` | チャンクの通し番号、0 から開始（単一 POST では常に 0）、必須 |
| `X-Total-Chunks` | 総チャンク数（1〜10。単一 POST では 1）、必須 |
| `X-Original-Url` | クライアントの元ページ URL。プレイヤー ID の解析に使用します（例：`https://.../user/123456...`）；**任意**、欠落時はプレイヤー ID が `unknown` になります |
| `X-Script-Version` | クライアントスクリプトのバージョン番号。サーバー側はこのヘッダーを無視するため送信不要 |

リクエストボディは生のバイナリセーブデータです（multipart 不要）。

制限：

- 単一ファイルの合計サイズ ≤1MB（`MAX_TOTAL_SIZE`）
- 単一チャンク ≤1MB（`MAX_CHUNK_SIZE`、超過すると 413 を返却）
- 総チャンク数 ≤10（`MAX_CHUNKS`）

> 注意：現在のセーブデータは約 200KB なので、**1 回の POST で送信完了できます**。分割アップロードは旧キャプチャクライアントとの互換性のために残しています。使用する場合は各チャンクを 1MB より十分に小さく（例：256KB）してください。10 チャンクで 1MB 上限まで送信できます。

レスポンス：

| ステータスコード | 意味 |
| --- | --- |
| `200` | セーブデータを受信し `OK` を返却。サーバー側が自動で完了します：セーブデータの結合（分割の場合）→ マップ生成 → `data/archive/by-id/<user_id>/<タイムスタンプ>/` へのアーカイブ → 通知のプッシュ。一切の手動介入は不要 |
| `400` | パラメータ不正（upload id の形式エラー、チャンク番号が範囲外、総チャンク数が 1〜10 の範囲外） |
| `413` | サイズ制限超過（単一チャンクが 1MB 超、または累計合計サイズが 1MB 超） |

### curl の例

単一 POST（現在のセーブデータは 1 回で送信完了できます）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

## Reqable レポートサーバー

カスタムキャプチャクライアントを使わずに、Reqable の「レポートサーバー」機能（v2.20.0+）を直接利用することもできます。キャプチャした各 HTTP セッションを [HAR](https://en.wikipedia.org/wiki/HAR_(file_format)) JSON 形式で自動的にあなたのサーバーへ POST し、gzip / brotli / zstd による圧縮も選択可能です。レポートエンドポイントは**デフォルトで有効**で、分割アップロードと共存します——`python cli.py server` で両方のエンドポイントが提供されます。無効にするには `REPORT_ENABLED=0` を設定します：

```bash
python cli.py server
```

設定（`.env`）：

| 変数 | デフォルト | 説明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1`（有効） | `0` / `false` を設定するとレポートエンドポイントが無効になります |
| `REPORT_PATH` | `/reqable/report` | エンドポイントのパス。Reqable の「アップロードパス」欄にこの値を入力します |
| `REPORT_MAX_SIZE` | `1` | HAR ボディのサイズ上限（MB。デフォルト 1、分割アップロードの上限と同じ） |
| `REPORT_TOKEN` | （空） | 任意の共有トークン。設定するとエンドポイントは `X-Report-Token` ヘッダーを要求します |

レポートを受け取るたびに、サーバーは次の処理を行います：

1. `Content-Encoding`（gzip / br / zstd）に従ってボディを展開し、HAR を解析します。
2. `log.entries` を走査し、「レスポンスボディ（フォールバック：リクエストボディ）が `AES_KEY` / `AES_IV` で復号でき、MySekai セーブデータとして解析できる」最初のセッションを採用します。ルールに一致してもセーブデータと無関係な通信はスキップされます。
3. セッション URL（`/user/<id>`、`X-Original-Url` と同じ規則）からプレイヤー ID を解決します。
4. セーブデータを `data/raw_mysekai/` に保存し、分割アップロードと同じ「生成 → アーカイブ → プッシュ」パイプラインを開始します。

注意：

- Reqable は各セッションを**1 回だけ送信し、失敗しても再試行しません**。そのためエンドポイントはできるだけ早く `200` を返します。サーバーを安定して稼働させ、`[REPORT]` ログに注意してください。
- 1 回のレポートにつき処理されるセーブデータは **1 つだけ**（最初の有効なエントリ）です。複数のエンドポイントに一致するルールでも、重複プッシュは発生しません。
- セキュリティ：プロトコル自体に認証はありません。Reqable はカスタムヘッダーを付加できないため、`REPORT_TOKEN` に頼るのではなく、`REPORT_PATH` にランダムな文字列を組み込む（例：`/reqable/report/9f3a…`）か、リバースプロキシ / ファイアウォールでアクセスを制限することをお勧めします。

Reqable 側の設定例：

- URL マッチングルール：`https://<ゲームAPIホスト>/api/user/*/mysekai*`
- アップロードパス：`http://<あなたのサーバー>:9478/reqable/report`
- 圧縮アルゴリズム：gzip / brotli / zstd のいずれでも可（サーバーは 3 種類すべて対応）

5 つのサーバーのゲーム API ホスト：

| サーバー | ゲーム API ホスト |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

推奨マッチングルール：`https://<ドメイン>/api/user/*/mysekai*`（CN で実測検証済み）。お使いのサーバーの mysekai API パスが異なる場合は、実際のパスに合わせてルールを調整してください。

手動 curl 検証（gzip 圧縮の HAR）：

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```

## プッシュの仕組み

### デフォルトは Telegram Bot

- `config/push_map.json` で設定されていないプレイヤーは、**すべてデフォルトで Telegram にプッシュ**されます。`push_map.json` ファイルが存在しない場合も同様に Telegram がデフォルトです。
- Telegram は Bot API の `sendMediaGroup` を使用し、4 枚のローカル PNG を multipart として直接アップロードします。**公開直リンクも静的ファイルサーバーも不要**です。`TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` が不足している場合は警告が出力されてスキップされ、Bark チャネルには影響しません。

### Bark プッシュは公開直リンクに依存

Bark（Day.app）通知の画像は **URL 直リンク**です：`notify.py` が画像アドレスを `image=` パラメータにエンコードして `api.day.app` へ送信し、Bark サーバーがその画像を取得します。そのため、この URL は**公開ネットワークから到達可能（HTTPS 推奨）**である必要があります。そうしないと Bark 通知に画像が含まれません。

4 枚のマップの直リンクは `notify.py` が以下の優先順位で組み立てます：

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| シナリオ | base の値 | 画像直リンクの形式 |
| --- | --- | --- |
| サーバーフロー（推奨） | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<タイムスタンプ>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<タイムスタンプ>/site_{5..8}.png` |
| 手動 CLI プッシュ | `BARK_IMAGE_BASE` または `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`（`data/latest/` を `<base>/` 配下に公開する必要があります） |

> 注意：サーバーフローでは `BARK_IMAGE_BASE` を設定した場合のみ、アーカイブパス付きの直リンクが組み立てられます。`FALLBACK_IMAGE_BASE` しか設定していない場合、サーバーがプッシュする直リンクも `<FALLBACK_IMAGE_BASE>/site_{5..8}.png` になります。

## 静的ファイルサーバーの例（任意）

目的：`data/archive/` ディレクトリを公開 URL として公開し、Bark サーバーが 4 枚のマップを取得できるようにします。

**推奨方法**：静的サーバーのルートディレクトリをプロジェクトの `data/` に向け、`BARK_IMAGE_BASE=https://<あなたのドメインまたはIP:ポート>` を設定すれば、自動的にマッピングされます：

```
data/archive/by-id/<user_id>/<タイムスタンプ>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<タイムスタンプ>/site_5.png
```

よく使う例：

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

注意事項：

- 直リンクのアドレスに **`127.0.0.1` / `localhost` は使わないでください**。Bark サーバーがそのアドレスにアクセスできる必要があるため、通常は公開ネットワークから到達可能なアドレスを選びます。内部ネットワークの IP は疎通を確認できた場合のみ使用します。
- **Telegram のみを使う場合は静的サーバーは一切不要**なので、この節は読み飛ばして構いません。
- 手動の `cli.py notify` の直リンクにはアーカイブパスが含まれないため、別途 `data/latest/` を `BARK_IMAGE_BASE` 配下に公開する必要があります。または `FALLBACK_IMAGE_BASE` を出力ディレクトリに向けます（例：`FALLBACK_IMAGE_BASE=http://<host>:5500/output` → そのサーバーが `data/latest/` を `/output` 配下にマウントします）。

## プレイヤープッシュルーティング（任意）

必要に応じて `config/` 配下にローカル設定を作成します（形式は同ディレクトリの `*.example.json` を参照。`.gitignore` で無視されています）：

- `push_map.json` — プレイヤー ID → プッシュ方法：値は `"telegram"`、Bark エイリアス、`"none"`（プッシュしない）。組み合わせ記法の `["alias", "telegram"]` や `"alias+tg"` にも対応しています。**未設定のプレイヤーはデフォルトで `telegram`** です。

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark エイリアス → デバイス key：

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## よくある質問

- **Bark 通知に画像が届かない？** 直リンクが公開ネットワークから到達可能か確認してください。ブラウザやモバイルネットワークで直接 `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<タイムスタンプ>/site_5.png` を開き、画像が表示されれば OK です。内部ネットワークのアドレス、`127.0.0.1`、または証明書が異常な HTTPS はいずれも画像取得に失敗します。
- **何もプッシュされない？** `push_map.json` でそのプレイヤーが `"none"` に設定されていないか確認してください。Bark のみを設定したユーザーが、そのプレイヤーに Bark エイリアスを設定し忘れていないか（未設定のプレイヤーはデフォルトで Telegram へ）。Telegram チャネルに token と chat id が設定されているか。Bark チャネルの key が不足していないか（`[BARK] ... failed` ログが出力されます）。
- **Bark は不要で Telegram だけ欲しい？** 何もする必要はありません——未設定のプレイヤーはデフォルトで Telegram に送られます。

## コマンドラインツール（cli.py）

すべての機能は `cli.py` から操作できます。インストール後（`pip install -e .`）は、同等の `mysekai` コマンドも使用できます。コマンドは成功時に終了コード 0、エラー時に 1 を返します（エラーメッセージは stderr に出力されます）。

```bash
python cli.py --help           # サブコマンド一覧
python cli.py <コマンド> --help     # 各サブコマンドの引数を確認
```

### generate —— セーブデータを復号してマップを生成

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`：暗号化セーブデータのパス（.bin）、必須
- フロー：AES-128-CBC 復号 → msgpack 解析 → ドロップ座標の抽出 → 4 枚のマップ描画（`site_5.png` ~ `site_8.png`）→ `rare_resources.txt` の出力
- `data/latest/` に出力し、最後に実際のパスを表示します
- 前提条件：`.env` に `AES_KEY` / `AES_IV` が設定されていること。セーブデータにドロップポイントが 1 つもない場合はエラーで終了します

### notify —— マップと統計をプッシュ

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`：`site_*.png` と `rare_resources.txt` を含むディレクトリ（通常は `data/latest/`）
- `[task_id]`：任意、アップロードタスク ID、デフォルトは `unknown`。`data/raw_mysekai/` からプレイヤー ID を逆引きするために使用します：`mysekai_<プレイヤーID>_<task_id>.bin` を優先的にマッチさせ、一致しない場合は raw_mysekai 内の最新セーブデータを使用します
- Telegram と Bark のどちらにプッシュするかは `config/push_map.json` のルーティングで決まります（未設定のプレイヤーはデフォルトで Telegram）。詳細は「プレイヤープッシュルーティング」を参照

### server —— アップロードサービスを起動（分割アップロード + Reqable レポートサーバー）

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- FastAPI サービスを起動します。クライアントは `POST /uploadMySekai` へ暗号化セーブデータをアップロード（単一 POST または分割。API の詳細は「アップロードAPI」を参照）、Reqable は HAR セッションを内蔵レポートエンドポイントへ報告できます（上記の「Reqable レポートサーバー」の章を参照）
- 全チャンク到着後、自動で完了します：セーブデータの結合 → マップ生成 → `data/archive/by-id/<user_id>/<タイムスタンプ>/` へのアーカイブ → プレイヤールーティングに従った通知のプッシュ。手動介入は不要
- デフォルトでは `9478` ポートをリッスンします。公開ネットワークにデプロイする場合は、リバースプロキシ経由で HTTPS として公開することを推奨します。クライアントスクリプトにハードコードされたアップロード URL（ポート含む）は、実際のデプロイ構成と一致させる必要があります

### 典型的な手動フロー

```bash
python cli.py generate mysekai_xxx.bin       # 1. data/latest/ にマップを生成
python cli.py notify data/latest <task_id>   # 2. プッシュ（task_id にはアップロード ID を入力、例：chfto53c3）
```

## ディレクトリ構成

```
├── app/                       # コアパッケージ
│   ├── config.py              # パス／環境変数／ローカル設定の集中管理
│   ├── crypto.py              # MySekai セーブデータの AES-128-CBC 復号
│   ├── parser.py              # msgpack 解析＋ステーション座標回転（純関数）
│   ├── har.py                 # Reqable レポートサーバー用 HAR 解析・解凍（純関数）
│   ├── render.py              # ドロップポイント抽出 → matplotlib 描画＋レア資源統計
│   ├── notify.py              # プッシュ：Telegram メディアグループ／Bark、プレイヤールーティング対応
│   ├── server.py              # FastAPI アップロードサービス（分割アップロード + Reqable レポートサーバー）
│   └── cli.py                 # コマンドラインエントリポイント
├── assets/                    # 静的リソース（リポジトリにコミット）
│   ├── resourceId.csv         # アイテム ID → 名称＋アイコン（base64）
│   └── NotoSansSC-Regular.ttf # 中国語フォント（OFL ライセンス）
├── config/                    # ローカル設定（実ファイルはコミットしない、*.example.json を参照）
│   ├── bark_map.example.json  # Bark エイリアス → デバイス key テンプレート
│   └── push_map.example.json  # プレイヤー ID → プッシュ方法テンプレート
├── data/                      # ランタイムデータ（ディレクトリ全体が gitignore）
│   ├── tmp/                   # 分割アップロードの一時保存、結合後にクリア
│   ├── raw_mysekai/           # 結合後の生（暗号化）セーブデータ、永続保持
│   ├── archive/               # 過去成果物のアーカイブ by-id/<user>/<タイムスタンプ>/（Bark 直リンクはここを指す）
│   └── latest/                # 直近に生成された成果物
├── cli.py                     # 統合エントリポイント
├── tests/                     # ユニットテスト（pytest）
├── .env.example               # 環境変数テンプレート（.env にコピーして記入）
└── requirements.txt           # ランタイム依存（バージョン固定）
```

## テスト

```bash
python -m pytest
```

## 免責事項

本ツールは個人の学習と娯楽のみを目的としています。商業用途やゲームの利用規約に違反する行為には使用しないでください。ゲームデータおよびアートリソースの著作権は元の権利者に帰属します。

## ライセンス

本プロジェクトのコードは [MIT License](LICENSE)（版权所有 © 2025 mouse233）を採用しています。自由に使用・変更・再配布できます。詳細は [LICENSE](LICENSE) を参照してください。

> ⚠️ ライセンスは本プロジェクトのコードのみを対象とします：`assets/` 内のゲーム素材（例：`resourceId.csv` 内のアイテムアイコン）とゲームデータの著作権は SEGA / Colorful Palette など元の権利者に帰属し、**MIT ライセンスの対象外**です。本ツール以外の用途に使用しないでください。
