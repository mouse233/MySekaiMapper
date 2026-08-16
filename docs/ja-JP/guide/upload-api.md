# アップロードAPI

クライアントはキャプチャした mysekai レスポンスボディを分割して `POST /uploadMySekai` に POST します（手動で curl を使い同じプロトコルでデバッグすることも可能です）。ヘッダーは以下の通りです：

| Header | 説明 |
| --- | --- |
| `X-Upload-Id` | アップロードタスク ID（英数字と `-` / `_` のみ、長さ 1〜64）、必須 |
| `X-Chunk-Index` | チャンクの通し番号、0 から開始、必須 |
| `X-Total-Chunks` | 総チャンク数（1〜10）、必須 |
| `X-Original-Url` | クライアントの元ページ URL。プレイヤー ID の解析に使用します（例：`https://.../user/123456...`）；**任意**、欠落時はプレイヤー ID が `unknown` になります |
| `X-Script-Version` | クライアントスクリプトのバージョン番号。サーバー側はこのヘッダーを無視するため送信不要 |

リクエストボディは生のバイナリチャンクデータです（multipart 不要）。

## 制限

- 単一ファイルの合計サイズ ≤1MB（`MAX_TOTAL_SIZE`）
- 単一チャンク ≤1MB（`MAX_CHUNK_SIZE`、超過すると 413 を返却）
- 総チャンク数 ≤10（`MAX_CHUNKS`）

::: tip
合計サイズの上限はわずか 1MB です。**チャンクサイズは 1MB より十分に小さくしないと意味がありません**（例：256KB なら 10 チャンクで 1MB いっぱいまで送信可能）。クライアントが 1MB チャンクを使用する場合、1MB を超えるファイルは 2 チャンク目以降 413 で拒否され、実質的に単一チャンクのみのアップロードに退化します。
:::

## レスポンス

| ステータスコード | 意味 |
| --- | --- |
| `200` | チャンクを受信し `OK` を返却。最後のチャンクが到着するとサーバー側が自動で完了します：セーブデータの結合 → マップ生成 → `data/archive/by-id/<user_id>/<タイムスタンプ>/` へのアーカイブ → 通知のプッシュ。一切の手動介入は不要 |
| `400` | パラメータ不正（upload id の形式エラー、チャンク番号が範囲外、総チャンク数が 1〜10 の範囲外） |
| `413` | サイズ制限超過（単一チャンクが 1MB 超、または累計合計サイズが 1MB 超） |

## curl の例

セーブデータが ≤1MB なら 1 チャンクで送信完了できます（最も一般的）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

分割アップロード（各チャンク 256KB、最大 10 チャンクで 1MB 上限まで送信）：

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
total=$(( ($(wc -c < "$file") + 262143) / 262144 ))
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_

i=0
for c in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H "X-Original-Url: https://example.com/user/1234567890123456789" \
    --data-binary @"$c"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

各チャンクが `200 OK` を返せば受信完了です。最後のチャンクが到着するとサーバー側が結合を開始し、以降のパイプラインは自動で完了します。`127.0.0.1:9478` は実際のサービスアドレスに置き換えてください。`X-Upload-Id` は `^[a-zA-Z0-9_-]{1,64}$` に一致する必要があります（例：`openssl rand -hex 5` で生成したランダム文字列）。
