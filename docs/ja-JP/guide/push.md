# プッシュの仕組み

## デフォルトは Telegram Bot

- `config/push_map.json` で設定されていないプレイヤーは、**すべてデフォルトで Telegram にプッシュ**されます。`push_map.json` ファイルが存在しない場合も同様に Telegram がデフォルトです。
- Telegram は Bot API の `sendMediaGroup` を使用し、4 枚のローカル PNG を multipart として直接アップロードします。**公開直リンクも静的ファイルサーバーも不要**です。`TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` が不足している場合は警告が出力されてスキップされ、Bark チャネルには影響しません。

## Bark プッシュは公開直リンクに依存

Bark（Day.app）通知の画像は **URL 直リンク**です：`notify.py` が画像アドレスを `image=` パラメータにエンコードして `api.day.app` へ送信し、Bark サーバーがその画像を取得します。そのため、この URL は**公開ネットワークから到達可能（HTTPS 推奨）**である必要があります。そうしないと Bark 通知に画像が含まれません。

4 枚のマップの直リンクは `notify.py` が以下の優先順位で組み立てます：

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| シナリオ | base の値 | 画像直リンクの形式 |
| --- | --- | --- |
| サーバーフロー（推奨） | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<タイムスタンプ>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<タイムスタンプ>/site_{5..8}.png` |
| 手動 CLI プッシュ | `BARK_IMAGE_BASE` または `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`（`data/latest/` を `<base>/` 配下に公開する必要があります） |

::: tip
サーバーフローでは `BARK_IMAGE_BASE` を設定した場合のみ、アーカイブパス付きの直リンクが組み立てられます。`FALLBACK_IMAGE_BASE` しか設定していない場合、サーバーがプッシュする直リンクも `<FALLBACK_IMAGE_BASE>/site_{5..8}.png` になります。
:::

画像を公開ネットワークへどう公開するかは[静的ファイルサーバー](/ja-JP/guide/static-server)を、プレイヤーを Telegram / Bark にどう割り当てるかは[プレイヤープッシュルーティング](/ja-JP/guide/routing)を参照してください。
