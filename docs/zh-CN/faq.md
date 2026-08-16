# 常见问题

- **Bark 通知收不到图片？** 检查直链是否公网可达：在浏览器/手机网络下直接打开 `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<时间戳>/site_5.png` 应能显示图片；内网地址、`127.0.0.1`、或证书异常的 HTTPS 都会导致抓图失败。
- **什么都没推送？** 检查 `push_map.json` 是否把该玩家设成了 `"none"`；只配了 Bark 的用户是否忘了在该玩家上配置 Bark 别名（未配置玩家默认走 Telegram）；Telegram 渠道是否配了 token 与 chat id；Bark 渠道是否缺 key（报 `[BARK] ... failed` 日志）。
- **不想收到 Bark 只想要 Telegram？** 什么都不用做——未配置的玩家默认就走 Telegram。
