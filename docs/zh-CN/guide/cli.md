# 命令行参考

先构建一次可执行文件：

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

如果二进制在仓库外执行，所有命令均可传入 `--root /path/to/MySekaiMapper`。

## inspect

```bash
bin/mysekaimapper inspect --input data/raw_mysekai/save.bin
```

解密并打印汇总，不生成图片。

## generate

```bash
bin/mysekaimapper generate --input data/raw_mysekai/save.bin --output data/latest
```

生成 `site_*.png` 与 `rare_resources.txt`；省略 `--output` 时默认写入 `data/latest`。

## notify

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789
```

`--player-id` 用于匹配 `config/push_map.json` 的玩家路由；`--image-base` 可覆盖 Bark 图片公开地址。

## serve

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

服务日志记录上传、队列、渲染、归档、通知、玩家 ID 与耗时；请按敏感运维数据处理日志。
