<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# MySekaiMapper

🌐 **Languages**: [English](../../) · [简体中文](../../zh-CN/) · [繁體中文](../../zh-TW/) · [日本語](../../ja-JP/) · [한국어](../../ko-KR/)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/ja-JP/>

暗号化された *Project SEKAI* の MySekai セーブデータを資源収集マップへ変換し、結果を Telegram または Bark（Day.app）へ送信する Go サービスです。

MitM キャプチャクライアントまたは Reqable の **Report Server** と連携します。キャプチャツールが MySekai セーブデータをアップロードすると、サービスが復号・解析してマップとレアリソース概要を描画し、成果物をアーカイブして、手動処理なしで通知を配信します。

通常の MySekai エリアでは `site_5.png`（草原）、`site_6.png`（浜辺）、`site_7.png`（花畑）、`site_8.png`（記念所）、および `rare_resources.txt` が生成されます。レンダラーと通知機能は、追加の通常 `site_*.png` 出力にも対応しています。

キャプチャフローは Nuverse が運営する中国（CN）および台湾（TW）サーバーで検証されています。他リージョンで利用できるかどうかは、その API パスとセーブデータ形式に依存します。

## 仕組み

```text
Game API response → MitM module / Reqable Report Server
    │  ① POST /uploadMySekai (single upload or ordered chunks)
    │  ② POST /reqable/report (HAR, optionally gzip / br / zstd)
    ▼
mysekaimapper serve
    ├─ AES-128-CBC decrypt + MsgPack parse + coordinate normalization
    ├─ render site_*.png + rare_resources.txt
    ├─ archive data/archive/by-id/<player_id>/<timestamp>/
    └─ publish data/latest/ and notify
         ├─ Telegram: upload local images as multipart media groups
         └─ Bark: send image URLs from a public static-file server
```
