<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# Go 重構

目前執行階段僅使用 Go。此模組採用標準根目錄結構，包含 `cmd/`、`internal/`、`go.mod` 與 `go.sum`；Python 原始碼、相依項目與 CI 均已移除。封存的參考實作仍保留在 [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) 分支與 [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) 標籤中。

HTTP 端點、環境變數、輸出名稱、封存配置與路由檔格式皆維持相容。Go 渲染器使用固定畫布，因此產生的 PNG 不保證與先前 Matplotlib 輸出逐像素相同。
