# FAQ

- **Bark notifications have no images?** Check whether the link is publicly reachable: open `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<timestamp>/site_5.png` in a browser or over cellular data — it should show the image. LAN addresses, `127.0.0.1`, or HTTPS with certificate problems all make the fetch fail.
- **Nothing was pushed?** Check whether `push_map.json` sets that player to `"none"`; whether Bark-only users forgot to assign a Bark alias to the player (unconfigured players default to Telegram); whether the Telegram channel has a token and chat id; whether the Bark channel lacks a key (look for `[BARK] ... failed` in the logs).
- **Don't want Bark, only Telegram?** Nothing to do — unconfigured players already default to Telegram.
