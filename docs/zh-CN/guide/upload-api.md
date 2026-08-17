# 上传接口

客户端把捕获的 mysekai 响应体通过 `POST /uploadMySekai` 上传（一次 POST 即可；分片上传仅作兼容保留）。手动用 curl 按同一协议调试亦可。header 如下：

| Header | 说明 |
| --- | --- |
| `X-Upload-Id` | 上传任务 ID（仅字母数字与 `-` / `_`，长度 1~64），必填 |
| `X-Chunk-Index` | 分片序号，从 0 开始（单片上传恒为 0），必填 |
| `X-Total-Chunks` | 总分片数（1~10；单片上传填 1），必填 |
| `X-Original-Url` | 客户端原始页面 URL，用于解析玩家 ID（如 `https://.../user/123456...`）；**可选**，缺失时玩家 ID 记为 `unknown` |
| `X-Script-Version` | 客户端脚本版本号；服务端忽略该头，可不传 |

请求体为原始二进制存档数据（无需 multipart）。

## 限制

- 单文件总大小 ≤1MB（`MAX_TOTAL_SIZE`）
- 单个分片 ≤1MB（`MAX_CHUNK_SIZE`，超限返回 413）
- 总分片数 ≤10（`MAX_CHUNKS`）

::: tip
当前存档约 200KB，**一次 POST 即可传完**。分片上传仅为兼容旧抓包客户端保留；若使用分片，每片应明显小于 1MB（例如 256KB），10 片可传满 1MB 上限。
:::

## 响应

| 状态码 | 含义 |
| --- | --- |
| `200` | 存档已接收，返回 `OK`；服务端自动完成：合并存档（如分片）→ 生成地图 → 归档到 `data/archive/by-id/<user_id>/<时间戳>/` → 推送通知，全程无需人工介入 |
| `400` | 参数非法（upload id 格式错误、分片序号越界、总分片数不在 1~10） |
| `413` | 超过大小限制（单分片超 1MB，或累计总大小超 1MB） |

## curl 示例

单次 POST（当前存档一次即可传完）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

分片上传（可选，兼容旧客户端；每片 256KB，10 片传满 1MB 上限）：

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

返回 `200 OK` 即表示存档已接收；流水线（如分片则先合并 → 生成地图 → 归档 → 推送）自动完成。把 `127.0.0.1:9478` 替换为你的实际服务地址；`X-Upload-Id` 必须匹配 `^[a-zA-Z0-9_-]{1,64}$`（例如用 `openssl rand -hex 5` 生成的随机串）。
