# Reqable 上报服务器

Reqable 可以将捕获到的 HAR 会话直接 POST 到 Go 服务。

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

默认端点为 `POST /reqable/report`，支持 `identity`、`gzip`、`br`、`zstd`、`zstandard`，也支持没有 content-size 字段的流式 zstd 帧。

环境变量：

- `REPORT_ENABLED=0`：关闭端点。
- `REPORT_PATH=/your/private/path`：修改端点路径。
- `REPORT_MAX_SIZE=1`：解压后的 HAR 请求体上限，单位 MiB。
- `REPORT_TOKEN`：要求匹配 `X-Report-Token` 请求头。

若 Reqable 无法附加令牌请求头，请使用随机私有路径并通过网络层 IP 白名单保护。
