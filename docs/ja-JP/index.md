---
layout: home

hero:
  name: MySekaiMapper
  text: MySekai 採集ポイントマップ生成ツール
  tagline: 「プロジェクトセカイ カラフルステージ！」（Project Sekai）MySekai 採集ポイントマップ生成・自動プッシュツール。
  actions:
    - theme: brand
      text: はじめる
      link: /ja-JP/guide/introduction
    - theme: alt
      text: GitHub で見る
      link: https://github.com/mouse233/MySekaiMapper

features:
  - title: 全自動パイプライン
    details: キャプチャツール（MitM モジュール / Reqable レポートサーバー）が mysekai データパケットを分割アップロードし、サーバー側が自動で結合・復号・描画・プッシュまで実行。一切の手動介入は不要です。
  - title: 4 枚のマップ + レア資源統計
    details: 1 回のタスクで site_5.png ~ site_8.png（さいしょの原っぱ、願いの砂浜、彩りの花畑、忘れ去られた場所）を生成し、さらに rare_resources.txt のレア資源統計も出力します。
  - title: Telegram 優先、Bark 対応
    details: Telegram は 4 枚の PNG を multipart で直接送信し、公開直リンクは不要。静的ファイルサーバーを設定すれば Bark にも画像直リンク通知が届きます。
  - title: AES-128-CBC 復号
    details: 暗号化された MySekai セーブデータを復号し、msgpack を解析してサイト座標を自動回転、matplotlib で採集マップを描画します。
---

## ワークフロー

```
ゲーム API 応答 → MitM モジュール / Reqable レポートサーバー（mysekai データをキャプチャ）
   │  ① 自動分割アップロード → server.py が自動結合（推奨、開発のきっかけ）
   │  ② または .bin セーブデータを手動配置 → cli.py generate
   ▼
parser.py    AES-128-CBC 復号 + msgpack 解析 + 座標回転
   ▼
render.py    site_5.png ~ site_8.png + rare_resources.txt を描画 → data/latest/
   ▼
notify.py    プッシュ：
             ├─ Telegram  ：画像を multipart で直接送信、公開直リンク不要 ← デフォルトチャネル
             └─ Bark      ：image= URL 直リンクで通知、静的ファイルサーバーが必要
```

> ⚠️ **免責事項**：本ツールは個人の学習と娯楽のみを目的としています。商業用途やゲームの利用規約に違反する行為には使用しないでください。ゲームデータおよびアートリソースの著作権は元の権利者に帰属します。

このサイトは以下の言語でもご覧いただけます：[English](/) · [简体中文](/zh-CN/) · [繁體中文](/zh-TW/) · [日本語](/ja-JP/) · [한국어](/ko-KR/)
