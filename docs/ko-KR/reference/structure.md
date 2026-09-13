<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 디렉터리 구조

```text
.
├── cmd/mysekaimapper/       # CLI entry point
├── internal/
│   ├── har/                 # Reqable HAR parsing and decompression
│   ├── mapper/              # AES, MsgPack, resources, and rendering
│   ├── notify/              # Telegram and Bark delivery
│   ├── server/              # Upload and report HTTP endpoints
│   └── service/             # Queue, storage, and archive pipeline
├── assets/                  # Font and resource icons
├── config/                  # Local routing templates
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # Ignored runtime data
│   ├── tmp/                 # Upload staging
│   ├── raw_mysekai/         # Encrypted source archives
│   ├── archive/             # Historical artifacts by player and timestamp
│   └── latest/              # Latest generated artifacts
├── docs/                    # VitePress documentation
├── go.mod / go.sum          # Go module definition
└── .env.example             # Configuration template
```

`data/`, `.env`, `config/bark_map.json` 및 `config/push_map.json`은 비공개 런타임 데이터이며 Git에서 무시됩니다.
