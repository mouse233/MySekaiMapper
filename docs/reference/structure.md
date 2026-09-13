<!-- GENERATED from README.md; do not edit directly. -->

# Directory structure

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

`data/`, `.env`, `config/bark_map.json`, and `config/push_map.json` are private runtime data and are ignored by Git.
