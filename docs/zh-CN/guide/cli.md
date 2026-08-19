<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 命令行参考

先构建一次可执行文件：

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

所有命令默认加载 `.env`，并接受 `--env /path/to/file`。`--root` 可放在子命令后的任意位置。

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

解密并解析存档，输出安全的聚合 JSON 摘要，不写入地图。

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

解密存档、提取掉落点，并写入 `site_*.png` 和 `rare_resources.txt`。`--output` 默认使用 `data/latest`；`--assets` 可覆盖资源目录。

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

`--output` 必填。`--task-id` 和 `--player-id` 默认值为 `unknown`；需要玩家专属路由时，请传入实际玩家 ID。

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

启动上传和上报 HTTP 端点，默认监听 `0.0.0.0:9478`。
