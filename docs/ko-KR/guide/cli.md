<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 명령줄 참조

바이너리를 한 번 빌드합니다.

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

모든 명령은 기본적으로 `.env`를 로드하며 `--env /path/to/file`을 받을 수 있습니다. `--root`는 하위 명령 뒤의 어느 위치에나 둘 수 있습니다.

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

저장 데이터를 복호화하고 파싱한 뒤, 지도를 작성하지 않고 안전한 집계 JSON 요약을 출력합니다.

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

아카이브를 복호화하고 드롭을 추출하여 `site_*.png`와 `rare_resources.txt`를 작성합니다. `--output`의 기본값은 `data/latest`이고, `--assets`로 에셋 디렉터리를 재정의할 수 있습니다.

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

`--output`은 필수입니다. `--task-id`와 `--player-id`의 기본값은 `unknown`이며, 플레이어별 라우팅이 필요한 경우 실제 플레이어 ID를 전달하세요.

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

업로드 및 보고서 HTTP 엔드포인트를 시작합니다. 기본값은 `0.0.0.0:9478`입니다.
