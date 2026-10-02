<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 알림 및 정적 파일

`config/push_map.example.json`과 `config/bark_map.example.json`에서 로컬 설정을 생성합니다. 이 파일에는 플레이어/기기 식별자가 포함되어 있으며 Git에서 무시됩니다.

### 플레이어 라우팅

`config/push_map.json`은 플레이어 ID를 `telegram`, `bark:<별칭>`, `astrbot:<별칭>`, `none`, 접두사가 있는 `+tg` 문자열 또는 메서드 배열에 매핑합니다:

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

사용 가능한 라우팅 값이 없는 플레이어는 기본적으로 Telegram을 사용합니다.

Bark 경로는 반드시 `bark:<별칭>`을 사용하고 AstrBot 경로는 `astrbot:<별칭>`을 사용합니다. 업데이트할 때 `push_map.json`의 `"sls"`, `"xufan"`을 `"bark:sls"`, `"bark:xufan"`으로 변경하세요. `bark_map.json` 키는 `"sls"`, `"xufan"` 그대로 유지합니다. 접두사가 없는 별칭은 경로 오류를 반환하지만 다른 유효한 대상은 계속 시도합니다. `"bark:klee+tg"` 문자열도 지원하지만 `["bark:klee", "telegram"]` 배열을 권장합니다.

### Telegram

Telegram은 생성된 모든 일반 `site_*.png` 파일을 로컬 multipart 미디어 그룹으로 업로드합니다. `TELEGRAM_BOT_TOKEN`과 `TELEGRAM_CHAT_ID`가 필요하지만, 공용 이미지 서버는 필요하지 않습니다. Telegram 전송이 실패해도 구성된 Bark 전송 시도는 중단되지 않습니다.

### AstrBot Push Lite (QQ)

기존 AstrBot에 [astrbot_plugin_push_lite](https://github.com/Raven95676/astrbot_plugin_push_lite)를 설치합니다. QQ는 OneBot v11로 연결하고 플러그인의 API token과 포트(기본 `9966`)를 설정합니다. `config/astrbot_map.example.json`에서 `config/astrbot_map.json`을 만듭니다:

```json
{
  "qq_me": {"platform_id": "qq_main", "type": "private", "qq": "123456789"},
  "qq_group": {"platform_id": "qq_main", "type": "group", "qq": "987654321"},
  "session": {"umo": "qq_main:FriendMessage:123456789"}
}
```

`platform_id`는 AstrBot의 실제 봇/플랫폼 인스턴스 ID이며 `aiocqhttp`나 봇 QQ 번호와 같다고 가정하면 안 됩니다. `type`은 `private` 또는 `group`, `qq`는 수신자 QQ 번호 또는 그룹 번호를 담은 **문자열**입니다. 프로그램이 `platform_id:FriendMessage:qq` 또는 `platform_id:GroupMessage:qq`를 생성합니다. `/sid`로 플랫폼 ID를 한 번 확인하면 수신자마다 명령을 실행할 필요가 없습니다. 전체 `umo`만 지정할 수도 있지만 다른 세 필드와 함께 사용할 수 없습니다. [AstrBot 세션 문서](https://docs.astrbot.app/use/command.html#name)를 참고하세요.

`config/push_map.json`에서 `astrbot:<별칭>`을 선택합니다. 기존 채널과 함께 사용할 수 있습니다:

```json
{
  "1234567890123456789": ["astrbot:qq_me", "astrbot:qq_group"],
  "1234567890123456790": ["astrbot:qq_me", "telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

`.env`를 설정합니다:

```dotenv
ASTRBOT_PUSH_URL=http://astrbot:9966
ASTRBOT_PUSH_TOKEN=your-plugin-api-token
ASTRBOT_ALLOW_INSECURE_HTTP=1
```

`ASTRBOT_PUSH_URL`은 `/send`를 제외한 서비스 **기본 URL**이며 역방향 프록시 경로 접두사를 지원합니다. 기본값은 HTTPS를 요구합니다. 신뢰할 수 있는 로컬/Docker 네트워크에서만 `ASTRBOT_ALLOW_INSECURE_HTTP=1`을 사용하세요. 별도 컨테이너에서는 공유 네트워크의 AstrBot 서비스 이름을 사용합니다. `localhost`는 현재 컨테이너를 가리킵니다. 호스트에서 실행한다면 접근 가능한 주소와 매핑한 플러그인 포트를 사용하세요. Bearer token으로 인증하며 리다이렉트를 따르지 않습니다.

`serve`와 `notify`는 각 수신자에게 텍스트 요약을 제출한 뒤 모든 일반 `site_*.png`를 개별 Base64 이미지로 제출합니다. 공개 이미지 서버나 공유 디렉터리가 필요하지 않습니다. 이미지당 인코딩 전 최대 8 MiB이며 프록시의 요청 크기 제한도 맞춰야 합니다. 같은 경로의 중복 AstrBot 별칭은 한 번만 제출합니다. 실패해도 다른 수신자와 채널을 계속 시도합니다.

성공 응답은 **Push Lite 대기열에 등록됨**을 뜻하며 QQ 전달 확인이 아닙니다. 로그에 `request queued`를 기록합니다. 이 연동은 전달 콜백이나 자동 재시도를 사용하지 않습니다. 플러그인의 메모리 대기열은 재시작 시 대기 메시지를 잃을 수 있으므로 최종 결과는 AstrBot 로그에서 확인하세요. 경로가 없는 플레이어의 기본값은 여전히 Telegram입니다. `.env`와 `config/astrbot_map.json`은 Git에서 무시됩니다.

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
