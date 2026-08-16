# 玩家推送路由（可选）

在 `config/` 下按需创建本地配置（格式见同目录 `*.example.json`，已被 `.gitignore` 忽略）：

- `push_map.json` — 玩家 ID → 推送方式：值为 `"telegram"`、Bark 别名、`"none"`（不推送），也支持组合写法 `["alias", "telegram"]` 或 `"alias+tg"`。**未配置的玩家默认 `telegram`**。

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark 别名 → 设备 key：

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 配置模板

| 文件 | 用途 | 模板 |
| --- | --- | --- |
| `config/push_map.json` | 玩家 ID → 推送方式路由 | `config/push_map.example.json` |
| `config/bark_map.json` | Bark 别名 → 设备 key | `config/bark_map.example.json` |
