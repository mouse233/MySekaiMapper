# Reqable レポートサーバー（任意）

Reqable の「レポートサーバー」機能（v2.20.0+）は、キャプチャした各 HTTP セッションを HAR JSON 形式で自動的にあなたのサーバーへ POST します。gzip / brotli / zstd による圧縮も選択可能です。サーバー側では `REPORT_ENABLED=1` で対応するエンドポイントを有効にします：

```bash
REPORT_ENABLED=1 python cli.py server
```

設定（`.env`）：

| 変数 | デフォルト | 説明 |
| --- | --- | --- |
| `REPORT_ENABLED` | （空 = 無効） | `1` / `true` を設定するとレポートエンドポイントが有効になります |
| `REPORT_PATH` | `/reqable/report` | エンドポイントのパス。Reqable の「アップロードパス」欄にこの値を入力します |
| `REPORT_MAX_SIZE` | `8` | HAR ボディのサイズ上限（MB）。セーブデータ自体は ≤1MB である必要があり、base64 で約 33% 膨張します |
| `REPORT_TOKEN` | （空） | 任意の共有トークン。設定するとエンドポイントは `X-Report-Token` ヘッダーを要求します |

## 処理の流れ

レポートを受け取るたびに、サーバーは次の処理を行います：

1. `Content-Encoding`（gzip / br / zstd）に従ってボディを展開し、HAR を解析します。
2. `log.entries` を走査し、「レスポンスボディ（フォールバック：リクエストボディ）が `AES_KEY` / `AES_IV` で復号でき、MySekai セーブデータとして解析できる」最初のセッションを採用します。ルールに一致してもセーブデータと無関係な通信はスキップされます。
3. セッション URL（`/user/<id>`）からプレイヤー ID を解決します。
4. セーブデータを `data/raw_mysekai/` に保存し、分割アップロードと同じ「生成 → アーカイブ → プッシュ」パイプラインを開始します。

::: warning
Reqable は各セッションを**1 回だけ送信し、失敗しても再試行しません**。そのためエンドポイントはできるだけ早く `200` を返します。サーバーを安定して稼働させ、`[REPORT]` ログに注意してください。
:::

1 回のレポートにつき処理されるセーブデータは **1 つだけ**（最初の有効なエントリ）です。複数のエンドポイントに一致するルールでも、重複プッシュは発生しません。

## セキュリティ

プロトコル自体に認証はありません。Reqable はカスタムヘッダーを付加できないため、`REPORT_TOKEN` に頼るのではなく、`REPORT_PATH` にランダムな文字列を組み込む（例：`/reqable/report/9f3a…`）か、リバースプロキシ / ファイアウォールでアクセスを制限することをお勧めします。

## Reqable 側の設定

- URL マッチングルール：`https://<ゲームAPIホスト>/*`（例：`https://<ゲームAPIホスト>/user/*/mysekai*` のように絞り込むことも可能）
- アップロードパス：`http://<あなたのサーバー>:9478/reqable/report`
- 圧縮アルゴリズム：gzip / brotli / zstd のいずれでも可（サーバーは 3 種類すべて対応）

## curl 例

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```
