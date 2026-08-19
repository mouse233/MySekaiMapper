<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# コマンドラインリファレンス

まずバイナリをビルドします。

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

すべてのコマンドはデフォルトで `.env` を読み込み、`--env /path/to/file` を受け付けます。`--root` はサブコマンドの後であればどこにでも指定できます。

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

セーブデータを復号・解析し、マップを書き出さずに安全な集計 JSON 概要を表示します。

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

アーカイブを復号してドロップを抽出し、`site_*.png` と `rare_resources.txt` を書き出します。`--output` の既定値は `data/latest` です。`--assets` でアセットディレクトリを上書きできます。

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

`--output` は必須です。`--task-id` と `--player-id` の既定値は `unknown` です。プレイヤー固有のルーティングが必要な場合は、必ず実際のプレイヤー ID を渡してください。

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

アップロードおよびレポート用 HTTP エンドポイントを起動します。既定値は `0.0.0.0:9478` です。
