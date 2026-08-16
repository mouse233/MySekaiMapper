# 上传接口

客户端把捕获的 mysekai 响应体分片 POST 到 `POST /uploadMySekai`（手动用 curl 按同一协议调试亦可）。header 如下：

| Header | 说明 |
| --- | --- |
| `X-Upload-Id` | 上传任务 ID（仅字母数字与 `-` / `_`，长度 1~64），必填 |
| `X-Chunk-Index` | 分片序号，从 0 开始，必填 |
| `X-Total-Chunks` | 总分片数（1~10），必填 |
| `X-Original-Url` | 客户端原始页面 URL，用于解析玩家 ID（如 `https://.../user/123456...`）；**可选**，缺失时玩家 ID 记为 `unknown` |
| `X-Script-Version` | 客户端脚本版本号；服务端忽略该头，可不传 |

请求体为原始二进制分片数据（无需 multipart）。

## 限制

- 单文件总大小 ≤1MB（`MAX_TOTAL_SIZE`）
- 单个分片 ≤1MB（`MAX_CHUNK_SIZE`，超限返回 413）
- 总分片数 ≤10（`MAX_CHUNKS`）

::: tip
总大小上限仅 1MB，**分片大小应明显小于 1MB 才有意义**（例如 256KB，10 片可传满 1MB）。若客户端用 1MB 分片，任何超过 1MB 的文件都会在第 2 片起被 413 拒绝，实际退化为只能单片上传。
:::

## 响应

| 状态码 | 含义 |
| --- | --- |
| `200` | 分片已接收，返回 `OK`；最后一片到达时服务端自动完成：合并存档 → 生成地图 → 归档到 `data/archive/by-id/<user_id>/<时间戳>/` → 推送通知，全程无需人工介入 |
| `400` | 参数非法（upload id 格式错误、分片序号越界、总分片数不在 1~10） |
| `413` | 超过大小限制（单分片超 1MB，或累计总大小超 1MB） |

## curl 示例

存档 ≤1MB 时单片即可传完（最常用）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

分片上传（每片 256KB，最多 10 片传满 1MB 上限）：

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

每个分片返回 `200 OK` 即表示已接收；最后一片到达后服务端开始合并，其余流水线自动完成。把 `127.0.0.1:9478` 替换为你的实际服务地址；`X-Upload-Id` 必须匹配 `^[a-zA-Z0-9_-]{1,64}$`（例如用 `openssl rand -hex 5` 生成的随机串）。
