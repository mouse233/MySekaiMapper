# 上传接口

`POST /uploadMySekai` 接收加密的 MySekai 存档，支持单次上传与有序分片上传。

| 请求头 | 含义 |
| --- | --- |
| `X-Upload-Id` | 1–64 字符的上传/任务 ID |
| `X-Chunk-Index` | 从零开始的分片序号 |
| `X-Total-Chunks` | 分片总数，最多 10 个 |
| `X-Original-Url` | 可选的游戏原始 URL；其中的 `/user/<id>` 用于玩家路由 |

加密存档默认上限为 1 MiB。成功响应为纯文本 `OK`；渲染与通知随后异步进行，请通过服务日志跟踪任务。
