<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 运行服务

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

服务会输出就绪地址，并记录上传/上报接收、入队、解析、渲染、归档、通知、耗时、任务 ID 与 `player_id` 等生命周期日志。日志不会记录存档正文、密钥、令牌或完整通知 URL。

进程处理 `SIGINT` 和 `SIGTERM`：先停止接收 HTTP 请求，再最多等待 15 秒处理已接收任务。

如果二进制在仓库外运行，请传入 `--root /path/to/MySekaiMapper`；否则会从工作目录自动发现仓库根目录。
