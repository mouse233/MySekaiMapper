<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 알림 및 정적 파일

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
