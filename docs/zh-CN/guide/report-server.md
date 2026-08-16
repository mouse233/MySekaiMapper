# Reqable 上报服务器

Reqable 内置的「上报服务器」功能（v2.20.0+）会把每个已捕获的 HTTP 会话按 HAR JSON 格式自动 POST 到你自建的服务端，可选 gzip / brotli / zstd 压缩。上报端点**默认开启**，与分片上传共存——`python cli.py server` 同时提供两个接口；设 `REPORT_ENABLED=0` 可关闭：

```bash
python cli.py server
```

配置（`.env`）：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1`（开启） | 设 `0` / `false` 关闭上报端点 |
| `REPORT_PATH` | `/reqable/report` | 端点路径，填入 Reqable 的「上报路径」 |
| `REPORT_MAX_SIZE` | `8` | HAR 请求体大小上限（MB）；存档本身需 ≤1MB，base64 膨胀约 33% |
| `REPORT_TOKEN` | （空） | 可选共享令牌；设置后端点要求请求头 `X-Report-Token` 匹配 |

## 处理流程

每次上报，服务端会：

1. 按 `Content-Encoding`（gzip / br / zstd）解压并解析 HAR。
2. 遍历 `log.entries`，取第一个「响应体（兜底：请求体）能用 `AES_KEY` / `AES_IV` 解密并解析为 MySekai 存档」的会话——命中规则但与存档无关的流量会被跳过。
3. 从会话 URL 解析玩家 ID（`/user/<id>`）。
4. 存档保存到 `data/raw_mysekai/`，并启动与分片上传相同的 生成 → 归档 → 推送 流水线。

::: warning
Reqable 每个会话**只上报 1 次且失败不重试**，因此端点会尽快返回 `200`。请保持服务稳定，并留意 `[REPORT]` 日志。
:::

每次上报只处理 **1 份**存档（第一个有效条目），因此匹配多个接口的规则不会造成重复推送。

## 安全

协议本身没有鉴权。Reqable 无法附加自定义请求头，建议把随机串拼进 `REPORT_PATH`（如 `/reqable/report/9f3a…`），或用反向代理 / 防火墙做访问限制，而不是依赖 `REPORT_TOKEN`。

## Reqable 侧配置

- URL 匹配规则：`https://<游戏API域名>/*`（或更精确，如 `https://<游戏API域名>/user/*/mysekai*`）
- 上报路径：`http://<你的服务器>:9478/reqable/report`
- 压缩算法：gzip / brotli / zstd 均可（服务端三种都支持）

## curl 示例

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```
