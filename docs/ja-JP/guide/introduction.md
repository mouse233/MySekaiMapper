# プロジェクト紹介

**MySekaiMapper** は「プロジェクトセカイ カラフルステージ！feat. 初音ミク」（Project Sekai）の MySekai（マイセカイ）採集ポイントマップ生成ツールです。

**開発のきっかけ**：MitM モジュールまたは Reqable の「レポートサーバー」機能と組み合わせて使用します——キャプチャツールがゲーム内の MySekai データパケットを捕捉した後、自動的に本サービスへアップロードします（1 回の POST で送信可能、分割アップロードにも対応）。サーバー側は暗号化セーブデータを復号し、各ステーションの資源ドロップ座標を抽出して採集マップを描画し、その結果（レア資源統計を含む）をプレイヤーの Telegram / Bark（iOS Day.app）へプッシュします。一切の手動介入は不要です。

1 回のタスクで **4 枚のマップ**が生成されます：`site_5.png`（さいしょの原っぱ）、`site_6.png`（願いの砂浜）、`site_7.png`（彩りの花畑）、`site_8.png`（忘れ去られた場所）。さらに `rare_resources.txt` のレア資源統計も出力します。

::: info サーバー互換性
本プロジェクトは朝夕光年（Nuverse）が運営する CN サーバー / TW サーバーで動作確認済みです。他のサーバーでの動作は未確認です。
:::

## ワークフロー

```
ゲーム API 応答 → MitM モジュール / Reqable レポートサーバー（mysekai データをキャプチャ）
   │  ① 自動アップロード（1 回の POST、分割にも対応）→ server.py が自動処理
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

## 環境要件

- Python 3.10+
- 実行時依存は `requirements.txt` を基準とします（バージョン固定）

## クイックナビゲーション

- [クイックスタート](/ja-JP/guide/quickstart) — インストール、`.env` の設定、パス A / パス B の選択
- [アップロードAPI](/ja-JP/guide/upload-api) — キャプチャクライアント向けの分割アップロードインターフェース
- [プッシュの仕組み](/ja-JP/guide/push) — Telegram / Bark 通知の仕組み
- [コマンドラインツール](/ja-JP/guide/cli) — `cli.py generate` / `notify` / `server`
