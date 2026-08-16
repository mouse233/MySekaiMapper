# Directory Structure

```
├── app/                       # core package
│   ├── config.py              # centralized paths / env vars / local config
│   ├── crypto.py              # MySekai save AES-128-CBC decryption
│   ├── parser.py              # msgpack parsing + station coordinate rotation (pure functions)
│   ├── render.py              # extract drop points → matplotlib drawing + rare-resource stats
│   ├── notify.py              # push: Telegram media groups / Bark, per-player routing
│   ├── server.py              # FastAPI chunked upload service
│   └── cli.py                 # CLI entry
├── assets/                    # static assets (committed to the repo)
│   ├── resourceId.csv         # item ID → name + icon (base64)
│   └── NotoSansSC-Regular.ttf # Chinese font (OFL license)
├── config/                    # local configs (real files not committed; see *.example.json)
│   ├── bark_map.example.json  # Bark alias → device key template
│   └── push_map.example.json  # player ID → push method template
├── data/                      # runtime data (whole directory gitignored)
│   ├── tmp/                   # chunk upload staging, cleaned after merge
│   ├── raw_mysekai/           # merged original (encrypted) saves, kept permanently
│   ├── archive/               # historical output archive by-id/<user>/<timestamp>/ (Bark links point here)
│   └── latest/                # most recent output
├── cli.py                     # unified entry
├── tests/                     # unit tests (pytest)
├── .env.example               # env var template (copy to .env and fill in)
└── requirements.txt           # runtime dependencies (pinned)
```

## Testing

```bash
python -m pytest
```
