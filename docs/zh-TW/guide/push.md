# 推播機制

## 預設走 Telegram Bot

- 未在 `config/push_map.json` 中設定的玩家，**一律預設推播到 Telegram**；`push_map.json` 檔案缺失時同樣預設 Telegram。
- Telegram 使用 Bot API `sendMediaGroup`，把 4 張本地 PNG 作為 multipart 直接上傳，**不需要公開網路直連，也不依賴靜態檔案伺服器**；`TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` 缺失時只印出警告並跳過，不影響 Bark 管道。

## Bark 推播依賴公開直連

Bark（Day.app）通知中的圖片是 **URL 直連**：`notify.py` 把圖片位址編碼進 `image=` 參數發給 `api.day.app`，由 Bark 伺服器再去抓取這張圖。因此該 URL 必須**公開網路可達（建議 HTTPS）**，否則 Bark 通知裡沒有圖片。

4 張地圖的直連由 `notify.py` 依以下優先順序拼出：

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| 場景 | base 取值 | 圖片直連形態 |
| --- | --- | --- |
| 伺服器流程（推薦） | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<時間戳>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<時間戳>/site_{5..8}.png` |
| 手動 CLI 推播 | `BARK_IMAGE_BASE` 或 `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`（需把 `data/latest/` 暴露在 `<base>/` 下） |

::: tip
伺服器流程只有在設定了 `BARK_IMAGE_BASE` 時才會拼出帶歸檔路徑的直連；若只設了 `FALLBACK_IMAGE_BASE`，伺服器推播的直連同樣是 `<FALLBACK_IMAGE_BASE>/site_{5..8}.png`。
:::

圖片如何暴露給公網見[靜態檔案伺服器](/zh-TW/guide/static-server)，玩家如何分配 Telegram / Bark 見[玩家推播路由](/zh-TW/guide/routing)。
