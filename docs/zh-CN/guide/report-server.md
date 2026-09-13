<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# Reqable 上报服务器

Reqable v2.20.0+ 可以将捕获的 HTTP 会话作为 HAR JSON POST 到本服务。上报端点默认启用，并与 `/uploadMySekai` 共存。

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | 设置为 `0`、`false`、`no` 或 `off` 可禁用上报 |
| `REPORT_PATH` | `/reqable/report` | 在 Reqable 中配置的端点路径 |
| `REPORT_MAX_SIZE` | `1` | 解压后的 HAR 请求体大小上限，单位为 MiB |
| `REPORT_TOKEN` | 空 | 可选；要求匹配 `X-Report-Token` |

### 处理流程

每次上报时，服务会：

1. 解压 `identity`、`gzip`、`br`、`zstd` 或 `zstandard` 内容并解析 HAR；也支持没有 content-size 字段的流式 zstd 帧。
2. 遍历 `log.entries`，接收第一个能够使用 `AES_KEY`/`AES_IV` 解密并验证为 MySekai 存档的响应体；响应体不匹配时回退检查请求体。
3. 从命中会话 URL 的 `/user/<id>` 中提取 `player_id`。
4. 将加密存档保存至 `data/raw_mysekai/`，并启动与上传接口相同的“绘制 → 归档 → 通知”流水线。

> Reqable 每个会话只上报一次，不会重试。请保持服务可用并关注 `[REPORT]` 日志。语法正确但不含 MySekai 存档的 HAR 仍会返回 `ok`；每次上报只处理第一份有效存档。

### Reqable 侧配置

- **匹配规则**：`https://<游戏 API 域名>/api/user/*/mysekai*`
- **服务器 URL**：`http://<你的服务器>:9478/reqable/report`（或自定义的 `REPORT_PATH`）

| 服务器 | 游戏 API 域名 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

该匹配规则已在国服验证。如果所在地区使用其他 MySekai API 路径，请检查实际抓包 URL 后调整规则。

### 安全

Reqable 无法附加自定义 `X-Report-Token` 请求头。请使用足够长的随机 `REPORT_PATH`，例如 `/reqable/report/<random>`，再通过反向代理或防火墙限制访问；不要在没有保护措施时公开默认端点。

### 手动 gzip HAR 测试

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```
