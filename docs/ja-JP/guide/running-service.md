<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# サービスの実行

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

サーバーは準備完了 URL を出力し、アップロード／レポートの受け付け、キュー投入、解析、描画、アーカイブ、通知、経過時間、タスク ID、`player_id` のライフサイクルログを書き出します。アーカイブ本文、シークレット、トークン、完全な通知 URL は意図的にログへ記録しません。

プロセスは `SIGINT` と `SIGTERM` を処理します。HTTP リクエストの受け付けを停止した後、すでに受け付けたジョブを最大 15 秒間処理します。

コンパイル済みバイナリは `--root /path/to/MySekaiMapper` を指定すればチェックアウト外から実行できます。指定しない場合は、作業ディレクトリからリポジトリルートを検出します。
