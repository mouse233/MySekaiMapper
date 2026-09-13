<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# Upload API

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
