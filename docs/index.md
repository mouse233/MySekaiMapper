---
layout: home

hero:
  name: MySekaiMapper
  text: MySekai gathering map generator
  tagline: A resource-gathering point map generator and auto-notifier for the MySekai mode in Project SEKAI.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/introduction
    - theme: alt
      text: View on GitHub
      link: https://github.com/mouse233/MySekaiMapper

features:
  - title: Zero-touch pipeline
    details: Capture tools (MitM module / Reqable Report Server) upload MySekai packets in chunks; the server merges, decrypts, draws maps, and pushes results automatically — no manual intervention.
  - title: 4 maps + rare stats
    details: Every task produces site_5.png ~ site_8.png (Empty Lot, Wish Beach, Flower Field, Place of Forgetting) plus a rare_resources.txt summary.
  - title: Telegram-first, Bark-ready
    details: Telegram receives the 4 PNGs via multipart directly — no public URL needed. Bark gets image links once you expose a static file server.
  - title: AES-128-CBC decryption
    details: Decrypts encrypted MySekai saves, parses msgpack, rotates station coordinates, and draws the gathering maps with matplotlib.
---

## How it works

```
Game API response → MitM module / Reqable Report Server (captures mysekai data)
   │  ① Auto chunked upload → server.py merges automatically (recommended; the original intent)
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

> ⚠️ **Disclaimer**: this tool is for personal learning and entertainment only. Do not use it for any commercial purpose or in ways that violate the game's terms of service. Game data and art assets belong to their respective owners.

The docs are available in [简体中文](/zh-CN/).
