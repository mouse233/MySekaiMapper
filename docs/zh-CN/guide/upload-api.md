<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 上传接口

`POST /uploadMySekai` 直接接收加密的 MySekai 响应正文。通常单次上传即可；为兼容抓包客户端，仍支持有序分片。

| 请求头 | 必填 | 含义 |
| --- | --- | --- |
| `X-Upload-Id` | 是 | 匹配 `^[A-Za-z0-9_-]{1,64}$` 的任务标识符 |
| `X-Chunk-Index` | 是 | 从零开始的分片序号 |
| `X-Total-Chunks` | 是 | 分片总数，范围为 1 到 10 |
| `X-Original-Url` | 否 | 游戏原始 URL；`/user/<id>` 用于提供玩家路由 |
| `X-Script-Version` | 否 | 为兼容抓包客户端而接受，服务会忽略 |

加密存档、每个分片和合并后的上传均限制为 1 MiB。成功接收后返回纯文本 `OK`；绘制和通知会在后台继续进行。

### 单次上传示例

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### 分片上传示例

使用相同的 `X-Upload-Id`、有序的索引，且最多十个分片：

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

常见响应包括：成功接收时为 `200 OK`，标识符或分片范围无效时为 `400 Bad Request`，超过大小限制时为 `413 Payload Too Large`，缺少必填请求头或其值不是整数时为 `422 Unprocessable Entity`。
