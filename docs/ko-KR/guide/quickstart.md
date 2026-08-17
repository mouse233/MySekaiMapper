# 빠른 시작

먼저 설치와 `.env` 기본 설정을 마친 뒤, 원하는 푸시 방식에 따라 경로를 선택하십시오:

- **경로 A(Telegram Bot 푸시만 사용)**: 설정이 가장 적으므로 먼저 이 경로를 통과시키는 것을 권장합니다.
- **경로 B(Bark 푸시 활성화)**: 경로 A를 기반으로 Bark key, 플레이어 라우팅, 정적 파일 서버를 추가로 설정해야 합니다.

## 1. 설치

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 선택 사항: mysekai 명령 설치 (python cli.py ... 와 동일)
venv/bin/pip install -e .
```

## 2. .env 설정(필수 항목)

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV`는 MySekai 저장 파일의 AES-128-CBC 복호화 키(각 16바이트)로, 어떤 경로를 선택하든 반드시 입력해야 합니다. 나머지 변수는 선택한 경로에 따라 설정합니다:

| 변수 | 필수 | 설명 |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | MySekai 저장 파일의 AES-128-CBC 키, 각 16바이트 |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | 선택* | Telegram 푸시(기본 채널)에 필요, [@BotFather](https://t.me/BotFather)에서 발급 |
| `BARK_ICON` | 선택 | Bark 알림 아이콘 URL |
| `BARK_IMAGE_BASE` | 선택 | 정적 파일 서버 루트 주소(Bark 이미지 직링크 푸시용, 아래 참조) |
| `FALLBACK_IMAGE_BASE` | 선택 | `BARK_IMAGE_BASE` 미설정 시 사용할 이미지 직링크 대체 주소 |

::: warning
\* Bark로만 알림을 받고 싶다면: Telegram 설정을 비워 둘 수 있지만, **`config/push_map.json`에서 플레이어를 Bark 별칭으로 라우팅해야 합니다**. 그렇지 않으면 미설정 플레이어는 기본적으로 Telegram으로 푸시되는데, Telegram 설정이 없으면 경고 한 줄만 출력하고 건너뛰므로 결국 아무것도 푸시되지 않습니다.
:::

## 3. 경로 A: Telegram Bot 푸시만 사용(가장 간단)

적용 시나리오: Telegram에서 지도와 통계만 받으면 되고, 다른 구성 요소는 건드리고 싶지 않은 경우.

1. `.env`에 Telegram 설정을 입력합니다([@BotFather](https://t.me/BotFather)에서 발급):

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 수동으로 한 번 실행해 검증합니다:

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. 일상 사용: 업로드 서버를 시작하면 세이브 도착 후 지도가 자동으로 생성되고 푸시됩니다. 캡처 방식은 두 가지입니다:

   - **MitM 모듈**: [업로드 API](/ko-KR/guide/upload-api)에 따라 업로드
   - **Reqable 보고서 서버**: 매칭 규칙과 업로드 경로 설정([Reqable 보고서 서버](/ko-KR/guide/report-server) 참조)

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

경로 A에서는 **필요하지 않습니다**: `config/push_map.json`, `config/bark_map.json`, 정적 파일 서버, `BARK_IMAGE_BASE`. 미설정 플레이어는 기본적으로 Telegram으로 푸시됩니다.

## 4. 경로 B: Bark 푸시 활성화(추가 설정 필요)

경로 A를 기반으로(Telegram 설정은 유지해도 되고, 비워 두고 Bark만 푸시해도 됩니다) 순서대로 다음을 채워 넣습니다:

1. **Bark key 설정**: `config/bark_map.json`에서 각 별칭에 대해 기기 key를 설정합니다(템플릿은 같은 디렉터리의 `bark_map.example.json` 참조).
2. **플레이어 라우팅 설정**: `config/push_map.json`에서 플레이어 ID를 Bark 별칭으로 라우팅합니다. 예:

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ::: warning
   **반드시 설정**: 미설정 플레이어는 기본적으로 Telegram으로 푸시됩니다. 이때 Telegram도 설정되어 있지 않으면 경고만 출력하고 건너뛰므로 결과적으로 아무것도 푸시되지 않습니다.
   :::
3. **정적 파일 서버 구축**: 프로젝트의 `data/` 디렉터리를 공개 네트워크에서 접근 가능한 HTTP(S) 서비스로 노출하고, `.env`에 `BARK_IMAGE_BASE=https://<도메인 또는 IP:포트>`를 설정합니다. 그렇지 않으면 Bark 알림에 지도 이미지가 포함되지 않습니다(자세한 내용은 [정적 파일 서버](/ko-KR/guide/static-server) 참조).
4. 검증과 일상 사용은 경로 A(2, 3단계)와 동일합니다.
