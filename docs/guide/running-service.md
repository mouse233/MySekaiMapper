<!-- GENERATED from README.md; do not edit directly. -->

# Running the service

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

The server prints ready URLs and writes lifecycle logs for upload/report acceptance, queueing, parsing, rendering, archiving, notifications, elapsed time, task ID, and `player_id`. It deliberately avoids logging archive bodies, secrets, tokens, or complete notification URLs.

The process handles `SIGINT` and `SIGTERM`: it stops accepting HTTP requests, then drains already accepted jobs for up to 15 seconds.

A compiled binary can run outside the checkout with `--root /path/to/MySekaiMapper`; otherwise the repository root is discovered from the working directory.
