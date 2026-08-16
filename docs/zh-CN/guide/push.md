# 推送机制

## 默认走 Telegram Bot

- 未在 `config/push_map.json` 中配置的玩家，**一律默认推送到 Telegram**；`push_map.json` 文件缺失时同样默认 Telegram。
- Telegram 使用 Bot API `sendMediaGroup`，把 4 张本地 PNG 作为 multipart 直接上传，**不需要公网直链，也不依赖静态文件服务器**；`TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` 缺失时只打印警告并跳过，不影响 Bark 渠道。

## Bark 推送依赖公开直链

Bark（Day.app）通知中的图片是 **URL 直链**：`notify.py` 把图片地址编码进 `image=` 参数发给 `api.day.app`，由 Bark 服务器再去抓取这张图。因此该 URL 必须**公网可达（建议 HTTPS）**，否则 Bark 通知里没有图片。

4 张地图的直链由 `notify.py` 按以下优先级拼出：

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| 场景 | base 取值 | 图片直链形态 |
| --- | --- | --- |
| 服务器流程（推荐） | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<时间戳>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<时间戳>/site_{5..8}.png` |
| 手动 CLI 推送 | `BARK_IMAGE_BASE` 或 `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`（需把 `data/latest/` 暴露在 `<base>/` 下） |

::: tip
服务器流程只有在配置了 `BARK_IMAGE_BASE` 时才会拼出带归档路径的直链；若只配了 `FALLBACK_IMAGE_BASE`，服务器推送的直链同样是 `<FALLBACK_IMAGE_BASE>/site_{5..8}.png`。
:::

图片如何暴露给公网见[静态文件服务器](/zh-CN/guide/static-server)，玩家如何分配 Telegram / Bark 见[玩家推送路由](/zh-CN/guide/routing)。
