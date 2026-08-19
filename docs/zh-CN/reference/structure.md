<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

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
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # 被忽略的运行时输出
│   ├── tmp/                 # 上传暂存
│   ├── raw_mysekai/         # 加密源存档
│   ├── archive/             # 按玩家和时间戳保存的历史产物
│   └── latest/              # 最新生成的产物
├── docs/                    # VitePress 文档
├── go.mod / go.sum          # Go 模块定义
└── .env.example             # 配置模板
```

`data/`、`.env`、`config/bark_map.json` 和 `config/push_map.json` 是私密的运行时数据，会被 Git 忽略。
