# ディレクトリ構成

```
├── app/                       # コアパッケージ
│   ├── config.py              # パス／環境変数／ローカル設定の集中管理
│   ├── crypto.py              # MySekai セーブデータの AES-128-CBC 復号
│   ├── parser.py              # msgpack 解析＋ステーション座標回転（純関数）
│   ├── har.py                 # Reqable レポートサーバー用 HAR 解析・解凍（純関数）
│   ├── render.py              # ドロップポイント抽出 → matplotlib 描画＋レア資源統計
│   ├── notify.py              # プッシュ：Telegram メディアグループ／Bark、プレイヤールーティング対応
│   ├── server.py              # FastAPI アップロードサービス（分割アップロード + Reqable レポートサーバー）
│   └── cli.py                 # コマンドラインエントリポイント
├── assets/                    # 静的リソース（リポジトリにコミット）
│   ├── resourceId.csv         # アイテム ID → 名称＋アイコン（base64）
│   └── NotoSansSC-Regular.ttf # 中国語フォント（OFL ライセンス）
├── config/                    # ローカル設定（実ファイルはコミットしない、*.example.json を参照）
│   ├── bark_map.example.json  # Bark エイリアス → デバイス key テンプレート
│   └── push_map.example.json  # プレイヤー ID → プッシュ方法テンプレート
├── data/                      # ランタイムデータ（ディレクトリ全体が gitignore）
│   ├── tmp/                   # 分割アップロードの一時保存、結合後にクリア
│   ├── raw_mysekai/           # 結合後の生（暗号化）セーブデータ、永続保持
│   ├── archive/               # 過去成果物のアーカイブ by-id/<user>/<タイムスタンプ>/（Bark 直リンクはここを指す）
│   └── latest/                # 直近に生成された成果物
├── cli.py                     # 統合エントリポイント
├── tests/                     # ユニットテスト（pytest）
├── .env.example               # 環境変数テンプレート（.env にコピーして記入）
└── requirements.txt           # ランタイム依存（バージョン固定）
```

## テスト

```bash
python -m pytest
```
