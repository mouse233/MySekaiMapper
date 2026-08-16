# 目录结构

```
├── app/                       # 核心包
│   ├── config.py              # 路径／环境变量／本地配置集中管理
│   ├── crypto.py              # MySekai 存档 AES-128-CBC 解密
│   ├── parser.py              # msgpack 解析＋站点坐标旋转（纯函数）
│   ├── render.py              # 提取掉落点 → matplotlib 绘图＋稀有资源统计
│   ├── notify.py              # 推送：Telegram 媒体组／Bark，按玩家路由
│   ├── server.py              # FastAPI 分片上传服务
│   └── cli.py                 # 命令行入口
├── assets/                    # 静态资源（提交到仓库）
│   ├── resourceId.csv         # 物品 ID → 名称＋图标（base64）
│   └── NotoSansSC-Regular.ttf # 中文字体（OFL 协议）
├── config/                    # 本地配置（真实文件不提交，参考 *.example.json）
│   ├── bark_map.example.json  # Bark 别名 → 设备 key 模板
│   └── push_map.example.json  # 玩家 ID → 推送方式模板
├── data/                      # 运行时数据（整个目录 gitignore）
│   ├── tmp/                   # 分片上传暂存，合并后即清
│   ├── raw_mysekai/           # 合并后的原始（加密）存档，永久保留
│   ├── archive/               # 历史成品归档 by-id/<user>/<时间戳>/（Bark 直链即指向此处）
│   └── latest/                # 最近一次生成的成品
├── cli.py                     # 统一入口
├── tests/                     # 单元测试（pytest）
├── .env.example               # 环境变量模板（复制为 .env 填写）
└── requirements.txt           # 运行时依赖（精确锁版本）
```

## 测试

```bash
python -m pytest
```
