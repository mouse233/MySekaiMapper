# 常見問題

- **Bark 通知收不到圖片？** 檢查直連是否公開網路可達：在瀏覽器/手機網路下直接開啟 `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<時間戳>/site_5.png` 應能顯示圖片；內網位址、`127.0.0.1`、或憑證異常的 HTTPS 都會導致抓圖失敗。
- **什麼都沒推播？** 檢查 `push_map.json` 是否把該玩家設成了 `"none"`；只配了 Bark 的使用者是否忘了在該玩家上設定 Bark 別名（未設定的玩家預設走 Telegram）；Telegram 管道是否配了 token 與 chat id；Bark 管道是否缺 key（報 `[BARK] ... failed` 日誌）。
- **不想收到 Bark 只想要 Telegram？** 什麼都不用做——未設定的玩家預設就走 Telegram。
