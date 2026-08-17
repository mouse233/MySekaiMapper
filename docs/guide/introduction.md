# Introduction

**MySekaiMapper** is a resource-gathering point map generator for the MySekai mode in *Project Sekai* (世界计划 多彩舞台).

**Original intent**: designed to work with MitM modules or Reqable's "Report Server" feature — the capture tool grabs MySekai data packets from the game and automatically uploads them to this service (single POST; chunked upload is also supported). The server decrypts the encrypted saves, extracts the resource drop coordinates of every station, draws gathering maps, and pushes the results (including a rare-resource summary) to the player's Telegram / Bark (iOS Day.app) — no manual intervention required.

Each task produces **4 maps**: `site_5.png` (Grassland), `site_6.png` (Beach), `site_7.png` (Flower Garden), `site_8.png` (Memorial Place), plus a `rare_resources.txt` rare-resource summary.

::: info Supported servers
This project has been tested and verified on the CN and TW servers operated by Nuverse (朝夕光年). Availability on other servers is unknown.
:::

## How it works

```
Game API response → MitM module / Reqable Report Server (captures mysekai data)
   │  ① Auto upload (single POST; chunked supported) → server.py processes automatically
   │  ② Or drop a .bin save manually → cli.py generate
   ▼
parser.py    AES-128-CBC decrypt + msgpack parse + coordinate rotation
   ▼
render.py    Draw site_5.png ~ site_8.png + rare_resources.txt → data/latest/
   ▼
notify.py    Push:
             ├─ Telegram: images uploaded directly as multipart, no public URL needed ← default channel
             └─ Bark: notified with image= URL links, requires a static file server
```

## Requirements

- Python 3.10+
- Dependencies pinned in `requirements.txt`

## Quick links

- [Quick Start](/guide/quickstart) — install, configure `.env`, pick Path A or Path B
- [Upload API](/guide/upload-api) — the chunked upload endpoint for capture clients
- [Push Mechanism](/guide/push) — how Telegram / Bark notifications work
- [CLI Reference](/guide/cli) — `cli.py generate` / `notify` / `server`
