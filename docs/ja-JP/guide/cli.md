# コマンドラインツール（cli.py）

すべての機能は `cli.py` から操作できます。インストール後（`pip install -e .`）は、同等の `mysekai` コマンドも使用できます。コマンドは成功時に終了コード 0、エラー時に 1 を返します（エラーメッセージは stderr に出力されます）。

```bash
python cli.py --help           # サブコマンド一覧
python cli.py <コマンド> --help     # 各サブコマンドの引数を確認
```

## generate —— セーブデータを復号してマップを生成

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`：暗号化セーブデータのパス（.bin）、必須
- フロー：AES-128-CBC 復号 → msgpack 解析 → ドロップ座標の抽出 → 4 枚のマップ描画（`site_5.png` ~ `site_8.png`）→ `rare_resources.txt` の出力
- `data/latest/` に出力し、最後に実際のパスを表示します
- 前提条件：`.env` に `AES_KEY` / `AES_IV` が設定されていること。セーブデータにドロップポイントが 1 つもない場合はエラーで終了します

## notify —— マップと統計をプッシュ

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`：`site_*.png` と `rare_resources.txt` を含むディレクトリ（通常は `data/latest/`）
- `[task_id]`：任意、アップロードタスク ID、デフォルトは `unknown`。`data/raw_mysekai/` からプレイヤー ID を逆引きするために使用します：`mysekai_<プレイヤーID>_<task_id>.bin` を優先的にマッチさせ、一致しない場合は raw_mysekai 内の最新セーブデータを使用します
- Telegram と Bark のどちらにプッシュするかは `config/push_map.json` のルーティングで決まります（未設定のプレイヤーはデフォルトで Telegram）。詳細は[プレイヤープッシュルーティング](/ja-JP/guide/routing)を参照

## server —— アップロードサービスを起動（分割アップロード + Reqable レポートサーバー）

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- FastAPI サービスを起動します。クライアントは `POST /uploadMySekai` へ暗号化セーブデータを分割アップロード（API の詳細は[アップロードAPI](/ja-JP/guide/upload-api)を参照）、Reqable は HAR セッションを内蔵レポートエンドポイントへ報告できます（[Reqable レポートサーバー](/ja-JP/guide/report-server)を参照）
- 全チャンク到着後、自動で完了します：セーブデータの結合 → マップ生成 → `data/archive/by-id/<user_id>/<タイムスタンプ>/` へのアーカイブ → プレイヤールーティングに従った通知のプッシュ。手動介入は不要
- デフォルトでは `9478` ポートをリッスンします。公開ネットワークにデプロイする場合は、リバースプロキシ経由で HTTPS として公開することを推奨します。クライアントスクリプトにハードコードされたアップロード URL（ポート含む）は、実際のデプロイ構成と一致させる必要があります

## 典型的な手動フロー

```bash
python cli.py generate mysekai_xxx.bin       # 1. data/latest/ にマップを生成
python cli.py notify data/latest <task_id>   # 2. プッシュ（task_id にはアップロード ID を入力、例：chfto53c3）
```
