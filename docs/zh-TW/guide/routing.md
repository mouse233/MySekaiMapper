# 玩家推播路由（選填）

在 `config/` 下依需求建立本地設定（格式見同目錄 `*.example.json`，已被 `.gitignore` 忽略）：

- `push_map.json` — 玩家 ID → 推播方式：值為 `"telegram"`、Bark 別名、`"none"`（不推播），也支援組合寫法 `["alias", "telegram"]` 或 `"alias+tg"`。**未設定的玩家預設 `telegram`**。

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark 別名 → 裝置 key：

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 設定範本

| 檔案 | 用途 | 範本 |
| --- | --- | --- |
| `config/push_map.json` | 玩家 ID → 推播方式路由 | `config/push_map.example.json` |
| `config/bark_map.json` | Bark 別名 → 裝置 key | `config/bark_map.example.json` |
