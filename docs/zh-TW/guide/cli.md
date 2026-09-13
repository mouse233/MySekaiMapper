<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 命令列參考

先建置一次二進位檔：

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

所有指令預設都會載入 `.env`，並接受 `--env /path/to/file`。`--root` 可放在子指令之後的任意位置。

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

解密並解析存檔，接著輸出安全的彙總 JSON 摘要，不會寫入地圖。

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

解密封存檔、擷取掉落物，並寫入 `site_*.png` 與 `rare_resources.txt`。`--output` 預設為 `data/latest`；可用 `--assets` 覆寫素材目錄。

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

必須提供 `--output`。`--task-id` 與 `--player-id` 預設為 `unknown`；需要依玩家進行路由時，請傳入實際的玩家 ID。

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

啟動上傳與報告 HTTP 端點。預設位址為 `0.0.0.0:9478`。
