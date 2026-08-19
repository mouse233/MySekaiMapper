# 目录结构

```text
.
├── cmd/mysekaimapper/       # CLI 入口
├── internal/
│   ├── har/                 # Reqable HAR 解析与解压
│   ├── mapper/              # AES、MsgPack、资源与渲染
│   ├── notify/              # Telegram 与 Bark 通知
│   ├── server/              # 上传与上报 HTTP 端点
│   └── service/             # 队列、存储与归档流水线
├── assets/                  # 字体和资源图标
├── config/                  # 本地路由模板
├── data/                    # 被忽略的运行时输出
├── docs/                    # VitePress 文档
├── go.mod / go.sum          # Go module 定义
└── .env.example             # 配置模板
```

`data/`、`.env` 与本地路由配置属于私密运行时数据，均不应提交到 Git。
