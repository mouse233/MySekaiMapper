<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# Reqable Report Server

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
