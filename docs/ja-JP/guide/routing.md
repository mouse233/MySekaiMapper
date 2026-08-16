# プレイヤープッシュルーティング（任意）

必要に応じて `config/` 配下にローカル設定を作成します（形式は同ディレクトリの `*.example.json` を参照。`.gitignore` で無視されています）：

- `push_map.json` — プレイヤー ID → プッシュ方法：値は `"telegram"`、Bark エイリアス、`"none"`（プッシュしない）。組み合わせ記法の `["alias", "telegram"]` や `"alias+tg"` にも対応しています。**未設定のプレイヤーはデフォルトで `telegram`** です。

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark エイリアス → デバイス key：

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 設定テンプレート

| ファイル | 用途 | テンプレート |
| --- | --- | --- |
| `config/push_map.json` | プレイヤー ID → プッシュ方法のルーティング | `config/push_map.example.json` |
| `config/bark_map.json` | Bark エイリアス → デバイス key | `config/bark_map.example.json` |
