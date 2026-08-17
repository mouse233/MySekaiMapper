# CLI Reference (cli.py)

Everything can be driven through `cli.py`; after installing (`pip install -e .`), the equivalent `mysekai` command is also available. Commands exit with 0 on success and 1 on error (errors print to stderr).

```bash
python cli.py --help           # subcommand overview
python cli.py <command> --help # show a subcommand's arguments
```

## generate — decrypt a save and generate maps

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`: path to the encrypted save (.bin), required
- Flow: AES-128-CBC decrypt → msgpack parse → extract drop coordinates → draw 4 maps (`site_5.png` ~ `site_8.png`) → write `rare_resources.txt`
- Output goes to `data/latest/`; the actual path is printed at the end
- Requirements: `AES_KEY` / `AES_IV` configured in `.env`; exits with an error if the save contains no drop points

## notify — push maps and stats

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`: directory containing `site_*.png` and `rare_resources.txt` (usually `data/latest/`)
- `[task_id]`: optional upload task ID, defaults to `unknown`. Used to look up the player ID from `data/raw_mysekai/`: it first tries to match `mysekai_<playerID>_<task_id>.bin`, otherwise falls back to the newest save in raw_mysekai
- Telegram vs Bark is decided by the routing in `config/push_map.json` (unconfigured players default to Telegram); see [Player Routing](/guide/routing)

## server — start the upload service (chunked upload + Reqable report server)

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- Starts the FastAPI service; clients upload encrypted saves to `POST /uploadMySekai` (single POST or chunked; protocol details: [Upload API](/guide/upload-api)), and Reqable can report HAR sessions to the built-in report endpoint (see [Reqable Report Server](/guide/report-server))
- When all chunks arrive, the server automatically: merges the save → generates maps → archives to `data/archive/by-id/<user_id>/<timestamp>/` → pushes notifications per player routing. No manual intervention.
- Listens on `9478` by default; for public deployment, expose it as HTTPS via a reverse proxy — the hardcoded upload URL (including the port) in your client script must match your actual deployment

## Typical manual flow

```bash
python cli.py generate mysekai_xxx.bin       # 1. generate maps to data/latest/
python cli.py notify data/latest <task_id>   # 2. push (task_id = upload ID, e.g. chfto53c3)
```
