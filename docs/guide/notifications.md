# Notifications

The service reads player routing from `config/push_map.json` and Bark device keys from `config/bark_map.json`. Both files are intentionally ignored by Git; start from their `.example.json` templates.

- **Telegram** uploads every generated regular `site_*.png` directly as media groups. A public image server is not required.
- **Bark** sends the rare-resource summary plus a notification for every generated regular `site_*.png`. Set `BARK_IMAGE_BASE` to the public base for archived maps, or use `--image-base` with the manual `notify` command.

Unconfigured players default to Telegram. The Go notifier skips symbolic links in the output directory and never logs tokens or complete request URLs.
