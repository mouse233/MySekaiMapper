<!-- GENERATED from doc/README.ja-JP.md; do not edit directly. -->

# Go リファクタリング

現在のランタイムは Go のみで構成されています。モジュールは `cmd/`、`internal/`、`go.mod`、`go.sum` による標準のルートレイアウトに従っています。Python のソース、依存関係、CI は削除されました。アーカイブ済みの参照実装は、[`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) ブランチおよび [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) タグに残されています。

HTTP エンドポイント、環境変数、出力名、アーカイブレイアウト、ルーティングファイル形式は互換性を維持しています。Go レンダラーは固定キャンバスを使用するため、生成される PNG が以前の Matplotlib 出力とピクセル単位で完全に同一になる保証はありません。
