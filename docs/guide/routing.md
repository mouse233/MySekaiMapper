# Player Push Routing (optional)

Create local configs under `config/` as needed (formats follow the `*.example.json` templates in the same directory; these files are `.gitignore`d):

- `push_map.json` — player ID → push method: the value can be `"telegram"`, a Bark alias, `"none"` (no push), or a combination like `["alias", "telegram"]` / `"alias+tg"`. **Unconfigured players default to `telegram`**.

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark alias → device key:

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## Configuration templates

| File | Purpose | Template |
| --- | --- | --- |
| `config/push_map.json` | player ID → push method routing | `config/push_map.example.json` |
| `config/bark_map.json` | Bark alias → device key | `config/bark_map.example.json` |
