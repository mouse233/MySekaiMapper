# 通知与路由

服务从 `config/push_map.json` 读取玩家路由，从 `config/bark_map.json` 读取 Bark 设备密钥。这两个文件均被 Git 忽略，请从对应的 `.example.json` 模板创建。

- **Telegram**：直接以媒体组上传全部生成的常规 `site_*.png`，不需要图片公网地址。
- **Bark**：发送稀有资源摘要，并为全部生成的常规 `site_*.png` 分别推送通知。设置 `BARK_IMAGE_BASE` 指向归档图片的公开根地址；手动 `notify` 可用 `--image-base` 覆盖。

未配置的玩家默认走 Telegram。Go 通知器会跳过输出目录中的符号链接，且不会记录 token 或完整请求 URL。
