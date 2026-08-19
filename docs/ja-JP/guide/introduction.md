<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# 仕組み

```text
Game API response → MitM module / Reqable Report Server
    │  ① POST /uploadMySekai (single upload or ordered chunks)
    │  ② POST /reqable/report (HAR, optionally gzip / br / zstd)
    ▼
mysekaimapper serve
    ├─ AES-128-CBC decrypt + MsgPack parse + coordinate normalization
    ├─ render site_*.png + rare_resources.txt
    ├─ archive data/archive/by-id/<player_id>/<timestamp>/
    └─ publish data/latest/ and notify
         ├─ Telegram: upload local images as multipart media groups
         └─ Bark: send image URLs from a public static-file server
```
