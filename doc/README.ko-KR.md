# MySekaiMapper

🌐 **Languages**: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/ko-KR/>

암호화된 *Project SEKAI* MySekai 저장 데이터를 자원 채집 지도으로 변환하고, 결과를 Telegram 또는 Bark(Day.app)로 전송하는 Go 서비스입니다.

MitM 캡처 클라이언트 또는 Reqable의 **Report Server**와 함께 사용할 수 있습니다. 캡처 도구가 MySekai 저장 데이터를 업로드하면, 서비스가 이를 복호화하고 파싱한 뒤 지도와 희귀 자원 요약을 렌더링하고 결과물을 보관하며, 수동 처리 없이 알림을 전송합니다.

일반적인 MySekai 지역에서는 `site_5.png`(초원), `site_6.png`(해변), `site_7.png`(꽃밭), `site_8.png`(기념 장소) 및 `rare_resources.txt`가 생성됩니다. 렌더러와 알림 전송기는 추가로 생성되는 일반 `site_*.png` 출력도 처리합니다.

캡처 흐름은 Nuverse가 운영하는 CN 및 TW 서버에서 검증되었습니다. 다른 지역에서의 사용 가능 여부는 해당 지역의 API 경로와 저장 데이터 형식에 따라 달라집니다.

## 작동 방식

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

## 빠른 시작

구성에 맞는 알림 경로를 선택하세요.

- **경로 A — Telegram만 사용**: 가장 간단한 선택지입니다. 플레이어 라우팅 파일이나 공개 이미지 서버가 필요하지 않습니다.
- **경로 B — Bark 사용**: Bark 키, 플레이어 라우팅 및 이미지용 공개 정적 파일 서버를 구성합니다.

### 1. 요구 사항 및 빌드

Go **1.25 이상**이 필요합니다.

```bash
go version
cp .env.example .env
go test ./...
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

`.env`의 `AES_KEY`와 `AES_IV`는 16바이트 AES-128-CBC 값으로 반드시 설정해야 합니다. `.env` 또는 로컬 라우팅 파일을 커밋하지 마세요.

### 2. `.env` 구성

| 변수 | 필수 여부 | 설명 |
| --- | --- | --- |
| `AES_KEY`, `AES_IV` | 예 | 16바이트 MySekai AES-128-CBC 키 및 IV |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` | Telegram 전용 | [@BotFather](https://t.me/BotFather)에서 받은 봇 자격 증명 및 대상 채팅 ID |
| `BARK_ICON` | 선택 사항 | Bark 알림에 포함할 아이콘 URL |
| `BARK_IMAGE_BASE` | Bark 이미지 사용 시 | 보관된 지도 이미지의 공개 기본 URL |
| `FALLBACK_IMAGE_BASE` | 선택 사항 | `BARK_IMAGE_BASE`가 설정되지 않았을 때 사용할 이미지 기본 URL |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | 선택 사항 | Reqable 보고서 엔드포인트 설정 |
| `MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, `MYSK_DATA_DIR` | 선택 사항 | 기본 저장소 디렉터리 재정의 |

### 3. 경로 A — Telegram만 사용

1. `.env`에 Telegram 변수를 설정합니다.

    ```dotenv
    TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
    TELEGRAM_CHAT_ID=123456789
    ```

2. 선택 사항으로 기존 암호화 저장 데이터로 파싱과 알림을 확인합니다.

    ```bash
    bin/mysekaimapper generate --input data/raw_mysekai/mysekai.bin
    bin/mysekaimapper notify \
      --output data/latest \
      --task-id manual-001 \
      --player-id 1234567890123456789
    ```

3. 일반 운영을 위해 서비스를 시작합니다.

    ```bash
    bin/mysekaimapper serve --host 0.0.0.0 --port 9478
    ```

`config/push_map.json`에 없는 플레이어는 기본적으로 Telegram으로 전송됩니다. 경로 A에는 Bark 맵, 푸시 맵 또는 공개 이미지 서버가 필요하지 않습니다.

### 4. 경로 B — Bark 사용

경로 A 구성에 더하여(오직 Bark로 라우팅하는 경우 Telegram은 생략 가능) 다음을 설정합니다.

1. `config/bark_map.example.json`을 바탕으로 `config/bark_map.json`을 만들고, 각 기기 키에 Bark 별칭을 매핑합니다.
2. `config/push_map.example.json`을 바탕으로 `config/push_map.json`을 만들고, 플레이어 ID를 Bark 별칭, `telegram`, `none` 또는 이들의 조합에 매핑합니다.

    ```json
    {
      "1234567890123456789": ["klee"],
      "1234567890123456790": ["telegram", "klee"],
      "1234567890123456791": "none"
    }
    ```

3. 저장소의 `data/` 디렉터리를 공개 HTTP(S) 정적 파일 서버로 노출하고, 그 공개 루트를 `BARK_IMAGE_BASE`로 설정합니다.

    ```dotenv
    BARK_IMAGE_BASE=https://maps.example.com
    ```

구성되지 않은 플레이어는 Telegram으로 전송됩니다. Telegram을 구성하지 않았다면 구성되지 않은 플레이어는 알림을 받지 않으므로, Bark만 사용하는 경우에는 Bark 별칭을 명시적으로 지정하세요.

## 서비스 실행

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

서버는 준비된 URL을 출력하고 업로드/보고서 수락, 큐 등록, 파싱, 렌더링, 보관, 알림, 경과 시간, 작업 ID 및 `player_id`에 대한 수명 주기 로그를 기록합니다. 보관 데이터 본문, 비밀값, 토큰 또는 전체 알림 URL은 의도적으로 로그에 남기지 않습니다.

프로세스는 `SIGINT` 및 `SIGTERM`을 처리합니다. HTTP 요청 수락을 중지한 뒤, 이미 수락된 작업을 최대 15초 동안 처리합니다.

컴파일된 바이너리는 `--root /path/to/MySekaiMapper`를 사용하여 체크아웃 외부에서 실행할 수 있습니다. 그렇지 않으면 작업 디렉터리를 기준으로 저장소 루트를 찾습니다.

## 업로드 API

`POST /uploadMySekai`는 암호화된 MySekai 응답 본문을 직접 받습니다. 일반적으로 단일 업로드만으로 충분하며, 캡처 클라이언트 호환성을 위해 순서가 있는 청크도 계속 지원합니다.

| 헤더 | 필수 여부 | 설명 |
| --- | --- | --- |
| `X-Upload-Id` | 예 | `^[A-Za-z0-9_-]{1,64}$`에 맞는 작업 식별자 |
| `X-Chunk-Index` | 예 | 0부터 시작하는 청크 인덱스 |
| `X-Total-Chunks` | 예 | 1에서 10 사이의 전체 청크 수 |
| `X-Original-Url` | 아니요 | 원본 게임 URL이며, `/user/<id>`로 플레이어 라우트를 판별 |
| `X-Script-Version` | 아니요 | 캡처 클라이언트 호환성을 위해 허용되며 서비스에서는 무시 |

암호화된 아카이브, 각 청크 및 병합된 업로드의 크기는 모두 1 MiB로 제한됩니다. 요청이 성공적으로 수락되면 일반 텍스트 `OK`를 반환하며, 렌더링과 알림은 백그라운드에서 계속됩니다.

### 단일 업로드 예시

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### 청크 업로드 예시

공통 `X-Upload-Id`, 순서대로 된 인덱스 및 최대 열 개의 청크를 사용하세요.

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_
total=$(ls /tmp/ms_chunk_* | wc -l | tr -d ' ')

i=0
for chunk in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
    --data-binary @"$chunk"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

일반적인 응답은 수락된 업로드에 대한 `200 OK`, 잘못된 식별자 또는 청크 범위에 대한 `400 Bad Request`, 크기 제한에 대한 `413 Payload Too Large`, 필수 업로드 헤더가 없거나 정수가 아닌 경우에 대한 `422 Unprocessable Entity`입니다.

## Reqable Report Server

Reqable v2.20.0 이상에서는 캡처한 각 HTTP 세션을 HAR JSON으로 이 서비스에 POST할 수 있습니다. 리포트 엔드포인트는 기본적으로 활성화되며 `/uploadMySekai`와 함께 사용할 수 있습니다.

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| 변수 | 기본값 | 설명 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | 리포트를 비활성화하려면 `0`, `false`, `no` 또는 `off`로 설정합니다 |
| `REPORT_PATH` | `/reqable/report` | Reqable에서 설정하는 엔드포인트 경로 |
| `REPORT_MAX_SIZE` | `1` | 압축 해제된 HAR 본문의 최대 크기(MiB) |
| `REPORT_TOKEN` | 비어 있음 | `X-Report-Token`에 필요한 선택적 값 |

### 처리 흐름

각 리포트에 대해 서비스는 다음을 수행합니다.

1. `identity`, `gzip`, `br`, `zstd` 또는 `zstandard` 콘텐츠의 압축을 해제하고 HAR을 파싱합니다. 콘텐츠 크기 필드가 없는 스트리밍 zstd 프레임도 지원합니다.
2. `log.entries`를 순회하며 `AES_KEY`/`AES_IV`로 복호화되고 MySekai 아카이브로 유효성이 검증되는 첫 번째 응답 본문을 수락합니다(응답 본문이 없으면 요청 본문으로 대체).
3. 일치하는 세션 URL의 `/user/<id>`에서 `player_id`를 추출합니다.
4. 암호화된 아카이브를 `data/raw_mysekai/`에 저장하고, 업로드에 사용되는 것과 동일한 render → archive → notify 파이프라인을 시작합니다.

> Reqable은 각 세션을 한 번만 리포트하며 재시도하지 않습니다. 서비스를 계속 사용할 수 있는 상태로 유지하고 `[REPORT]` 로그를 확인하세요. MySekai 아카이브가 없는 구문상 유효한 HAR도 `ok`를 반환합니다. 리포트에서 처리되는 것은 첫 번째 유효한 아카이브뿐입니다.

### Reqable 구성

- **매칭 규칙**: `https://<game-api-domain>/api/user/*/mysekai*`
- **서버 URL**: `http://<your-server>:9478/reqable/report`(또는 사용자 지정 `REPORT_PATH`)

| 서버 | 게임 API 도메인 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

이 매칭 패턴은 CN에서 검증되었습니다. 해당 지역에서 다른 MySekai API 경로를 사용하는 경우 캡처된 URL을 확인하고 규칙을 조정하세요.

### 보안

Reqable은 사용자 지정 `X-Report-Token` 헤더를 추가할 수 없습니다. `/reqable/report/<random>`과 같이 길고 무작위적인 `REPORT_PATH`를 사용하고, 리버스 프록시 또는 방화벽을 통해 접근을 제한하세요. 별도의 제어 없이 기본 엔드포인트를 외부에 공개하지 마세요.

### 수동 gzip HAR 테스트

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```

## 알림 및 정적 파일

`config/push_map.example.json`과 `config/bark_map.example.json`에서 로컬 설정을 생성합니다. 이 파일에는 플레이어/기기 식별자가 포함되어 있으며 Git에서 무시됩니다.

### 플레이어 라우팅

`config/push_map.json`은 플레이어 ID를 `telegram`, Bark 별칭, `none`, `+tg` 문자열 또는 메서드 배열에 매핑합니다:

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "klee"],
  "1234567890123456791": "none"
}
```

사용 가능한 라우팅 값이 없는 플레이어는 기본적으로 Telegram을 사용합니다.

### Telegram

Telegram은 생성된 모든 일반 `site_*.png` 파일을 로컬 multipart 미디어 그룹으로 업로드합니다. `TELEGRAM_BOT_TOKEN`과 `TELEGRAM_CHAT_ID`가 필요하지만, 공용 이미지 서버는 필요하지 않습니다. Telegram 전송이 실패해도 구성된 Bark 전송 시도는 중단되지 않습니다.

### Bark

Bark는 희귀 리소스 요약을 전송하고, 생성된 모든 일반 `site_*.png` 파일에 대해 개별적으로 알림을 보냅니다. `config/bark_map.json`은 별칭과 기기 키의 매핑을 담당합니다:

```json
{ "klee": "paste-your-bark-key-here" }
```

Bark는 이미지 URL을 직접 가져옵니다. 자동 서비스 작업에서는 `BARK_IMAGE_BASE`가 공개된 `data/` 루트를 가리키도록 설정해야 하며, 아카이브 URL은 다음과 같습니다:

```text
https://maps.example.com/archive/by-id/<player_id>/<timestamp>/site_5.png
```

수동 `notify`의 이미지 루트 경로 우선순위는 `--image-base`, `BARK_IMAGE_BASE`, `FALLBACK_IMAGE_BASE`입니다. 이 루트 경로에서는 선택한 출력 디렉터리를 직접 공개해야 합니다.

### 정적 파일 서버

Bark 이미지에는 `localhost`나 `127.0.0.1`을 사용할 수 없습니다. 다음과 같이 공용 HTTPS를 사용하세요:

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    root /path/to/MySekaiMapper/data;
}
```

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

알림기는 출력 디렉터리의 심볼릭 링크를 무시하며, 자격 증명이나 전체 알림 URL을 기록하지 않습니다.

## 명령줄 참조

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

## 디렉터리 구조

```text
.
├── cmd/mysekaimapper/       # CLI entry point
├── internal/
│   ├── har/                 # Reqable HAR parsing and decompression
│   ├── mapper/              # AES, MsgPack, resources, and rendering
│   ├── notify/              # Telegram and Bark delivery
│   ├── server/              # Upload and report HTTP endpoints
│   └── service/             # Queue, storage, and archive pipeline
├── assets/                  # Font and resource icons
├── config/                  # Local routing templates
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # Ignored runtime data
│   ├── tmp/                 # Upload staging
│   ├── raw_mysekai/         # Encrypted source archives
│   ├── archive/             # Historical artifacts by player and timestamp
│   └── latest/              # Latest generated artifacts
├── docs/                    # VitePress documentation
├── go.mod / go.sum          # Go module definition
└── .env.example             # Configuration template
```

`data/`, `.env`, `config/bark_map.json` 및 `config/push_map.json`은 비공개 런타임 데이터이며 Git에서 무시됩니다.

## 테스트

```bash
go test ./...
go build -o /tmp/mysekaimapper ./cmd/mysekaimapper
npm run docs:build
```

GitHub Actions는 푸시와 풀 리퀘스트에 대해 Go 테스트 모음과 빌드를 실행합니다.

## Go 리팩터링

현재 런타임은 Go만 사용합니다. 모듈은 `cmd/`, `internal/`, `go.mod`, `go.sum`으로 이루어진 표준 루트 구조를 따르며 Python 소스, 의존성 및 CI는 제거되었습니다. 보관된 참조 구현은 [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) 브랜치와 [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) 태그에 남아 있습니다.

HTTP 엔드포인트, 환경 변수, 출력 이름, 아카이브 레이아웃 및 라우팅 파일 형식은 호환성을 유지합니다. Go 렌더러는 고정 캔버스를 사용하므로 생성되는 PNG가 이전 Matplotlib 출력과 픽셀 단위로 동일하다고 보장되지는 않습니다.

## 면책 조항

이 도구는 개인 학습 및 오락 목적으로만 사용해야 합니다. 상업적 목적으로 사용하거나 게임의 서비스 약관을 위반하는 방식으로 사용하지 마세요. 게임 데이터와 에셋은 각각의 소유자에게 귀속됩니다.

## 라이선스

프로젝트 코드는 [MIT](https://github.com/mouse233/MySekaiMapper/blob/feat/go-rewrite/LICENSE) 라이선스로 제공됩니다(Copyright © 2025 mouse233). `assets/` 아래의 게임 에셋과 게임 데이터는 각각의 소유자에게 귀속되며 이 라이선스의 적용 대상이 아닙니다.
