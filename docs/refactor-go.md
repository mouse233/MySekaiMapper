<!-- GENERATED from README.md; do not edit directly. -->

# Go refactor

The active runtime is Go-only. The module follows the standard root layout with `cmd/`, `internal/`, `go.mod`, and `go.sum`; Python source, dependencies, and CI were removed. The archived reference implementation remains in the [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) branch and [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) tag.

The HTTP endpoints, environment variables, output names, archive layout, and routing-file formats remain compatible. The Go renderer uses a fixed canvas, so its generated PNGs are not guaranteed to be pixel-identical to the former Matplotlib output.
