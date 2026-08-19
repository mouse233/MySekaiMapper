<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 빠른 시작

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
