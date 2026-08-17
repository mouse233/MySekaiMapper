# アップロードAPI

クライアントはキャプチャした mysekai レスポンスボディを `POST /uploadMySekai` にアップロードします（1 回の POST で送信可能。分割アップロードは互換性のため残しています）。手動で curl を使い同じプロトコルでデバッグすることも可能です。ヘッダーは以下の通りです：

| Header | 説明 |
| --- | --- |
| `X-Upload-Id` | アップロードタスク ID（英数字と `-` / `_` のみ、長さ 1〜64）、必須 |
| `X-Chunk-Index` | チャンクの通し番号、0 から開始（単一 POST では常に 0）、必須 |
| `X-Total-Chunks` | 総チャンク数（1〜10。単一 POST では 1）、必須 |
| `X-Original-Url` | クライアントの元ページ URL。プレイヤー ID の解析に使用します（例：`https://.../user/123456...`）；**任意**、欠落時はプレイヤー ID が `unknown` になります |
| `X-Script-Version` | クライアントスクリプトのバージョン番号。サーバー側はこのヘッダーを無視するため送信不要 |

リクエストボディは生のバイナリセーブデータです（multipart 不要）。

## 制限

- 単一ファイルの合計サイズ ≤1MB（`MAX_TOTAL_SIZE`）
- 単一チャンク ≤1MB（`MAX_CHUNK_SIZE`、超過すると 413 を返却）
- 総チャンク数 ≤10（`MAX_CHUNKS`）

::: tip
現在のセーブデータは約 200KB なので、**1 回の POST で送信完了できます**。分割アップロードは旧キャプチャクライアントとの互換性のために残しています。使用する場合は各チャンクを 1MB より十分に小さく（例：256KB）してください。10 チャンクで 1MB 上限まで送信できます。
:::

## レスポンス

| ステータスコード | 意味 |
| --- | --- |
| `200` | セーブデータを受信し `OK` を返却。サーバー側が自動で完了します：セーブデータの結合（分割の場合）→ マップ生成 → `data/archive/by-id/<user_id>/<タイムスタンプ>/` へのアーカイブ → 通知のプッシュ。一切の手動介入は不要 |
| `400` | パラメータ不正（upload id の形式エラー、チャンク番号が範囲外、総チャンク数が 1〜10 の範囲外） |
| `413` | サイズ制限超過（単一チャンクが 1MB 超、または累計合計サイズが 1MB 超） |

## curl の例

単一 POST（現在のセーブデータは 1 回で送信完了できます）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

分割アップロード（任意、旧クライアントとの互換用。各チャンク 256KB、10 チャンクで 1MB 上限まで送信）：

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

`200 OK` が返ればセーブデータは受信完了です。以降のパイプライン（分割の場合は結合 → マップ生成 → アーカイブ → プッシュ）は自動で完了します。`127.0.0.1:9478` は実際のサービスアドレスに置き換えてください。`X-Upload-Id` は `^[a-zA-Z0-9_-]{1,64}$` に一致する必要があります（例：`openssl rand -hex 5` で生成したランダム文字列）。
