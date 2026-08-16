# クイックスタート

まずインストールと `.env` の基本設定を済ませてから、希望するプッシュ方法に応じてパスを選択してください：

- **パス A（Telegram Bot プッシュのみ）**：設定が最も少なく、まずこちらで一通り動かすことをお勧めします。
- **パス B（Bark プッシュを有効化）**：パス A に加えて、Bark key、プレイヤールーティング、静的ファイルサーバーの追加設定が必要です。

## 1. インストール

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 任意:mysekai コマンドをインストール(python cli.py ... と同等)
venv/bin/pip install -e .
```

## 2. .env の設定（必須項目）

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

::: warning
\* Bark だけで通知を受け取りたい場合：Telegram の設定は空のままで構いませんが、**`config/push_map.json` でプレイヤーを Bark エイリアスにルーティングする必要があります**。そうしないと、未設定のプレイヤーはデフォルトで Telegram へ送られますが、Telegram の設定が不足している場合は警告が 1 行出力されてスキップされ、結果として何もプッシュされません。
:::

## 3. パス A：Telegram Bot プッシュのみ（最小構成）

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

3. 日常利用：アップロードサーバーを起動します。キャプチャクライアント（MitM モジュール / Reqable レポートサーバー）が[アップロードAPI](/ja-JP/guide/upload-api)に従って分割アップロードすると、自動的にマップが生成されプッシュされます：

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

パス A で**不要なもの**：`config/push_map.json`、`config/bark_map.json`、静的ファイルサーバー、`BARK_IMAGE_BASE`。未設定のプレイヤーはデフォルトで Telegram にプッシュされます。

## 4. パス B：Bark プッシュを有効化（追加設定が必要）

パス A をベースに（Telegram の設定はそのままでも、空のままにして Bark のみでも可）、以下の順に設定を追加します：

1. **Bark key の設定**：`config/bark_map.json` で各エイリアスにデバイス key を設定します（テンプレートは同ディレクトリの `bark_map.example.json` を参照）。
2. **プレイヤールーティングの設定**：`config/push_map.json` でプレイヤー ID を Bark エイリアスにルーティングします。例：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ::: warning
   **必須設定**：未設定のプレイヤーはデフォルトで Telegram に送られます。その時点で Telegram も未設定の場合は、警告が出力されてスキップされ、結果として何もプッシュされません。
   :::
3. **静的ファイルサーバーの構築**：プロジェクトの `data/` ディレクトリを公開ネットワークからアクセス可能な HTTP(S) サービスとして公開し、`.env` に `BARK_IMAGE_BASE=https://<ドメインまたはIP:ポート>` を設定します。そうしないと Bark 通知にマップ画像が含まれません（詳細は[静的ファイルサーバー](/ja-JP/guide/static-server)を参照）。
4. 検証と日常利用はパス A（手順 2、3）と同じです。
