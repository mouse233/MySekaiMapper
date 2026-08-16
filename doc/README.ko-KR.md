# MySekaiMapper

🌐 언어: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/ko-KR/>

프로젝트 세카이 컬러풀 스테이지! feat. 하츠네 미쿠(Project Sekai)의 MySekai(내 세카이) 수집 포인트 지도 생성 도구입니다.

**프로젝트 취지**: MitM 모듈 또는 Reqable의「보고서 서버」기능과 함께 사용합니다——패킷 캡처 도구가 게임 내 MySekai 데이터 패킷을 캡처하면 자동으로 이 서비스에 분할 업로드하고, 서버가 암호화된 저장 파일을 병합·복호화하여 각 사이트의 자원 드롭 좌표를 추출한 뒤 수집 지도를 그려 그 결과(희귀 자원 통계 포함)를 플레이어의 Telegram / Bark(iOS Day.app)로 푸시합니다. 전 과정에서 수동 개입이 필요 없습니다.

한 번의 작업으로 **4장의 지도**가 생성됩니다: `site_5.png`(시작의 공터), `site_6.png`(소원의 해변), `site_7.png`(화려한 꽃밭), `site_8.png`(잊혀진 곳), 그리고 `rare_resources.txt` 희귀 자원 통계 1개가 추가로 생성됩니다.

본 프로젝트는 Nuverse(조석광년)가 운영하는 CN 서버 / TW 서버에서 테스트를 통과했습니다. 다른 서버에서의 동작은 확인되지 않았습니다.

## 작업 흐름

```
게임 API 응답 → MitM 모듈 / Reqable 보고서 서버(패킷 캡처로 mysekai 데이터 확보)
   │  ① 자동 분할 업로드 → server.py 자동 병합(권장, 프로젝트 취지)
   │  ② 또는 .bin 저장 파일 직접 배치 → cli.py generate
   ▼
parser.py    AES-128-CBC 복호화 + msgpack 파싱 + 좌표 회전
   ▼
render.py    site_5.png ~ site_8.png + rare_resources.txt 생성 → data/latest/
   ▼
notify.py    푸시:
             ├─ Telegram  : 이미지 multipart 직접 전송, 공개 직링크 불필요 ← 기본 채널
             └─ Bark      : image= URL 직링크 알림, 정적 파일 서버 필요
```

## 빠른 시작

먼저 설치와 `.env` 기본 설정을 마친 뒤, 원하는 푸시 방식에 따라 경로를 선택하십시오:

- **경로 A(Telegram Bot 푸시만 사용)**: 설정이 가장 적으므로 먼저 이 경로를 통과시키는 것을 권장합니다.
- **경로 B(Bark 푸시 활성화)**: 경로 A를 기반으로 Bark key, 플레이어 라우팅, 정적 파일 서버를 추가로 설정해야 합니다.

### 1. 설치

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 선택 사항: mysekai 명령 설치 (python cli.py ... 와 동일)
venv/bin/pip install -e .
```

### 2. .env 설정(필수 항목)

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

> \* Bark로만 알림을 받고 싶다면: Telegram 설정을 비워 둘 수 있지만, **`config/push_map.json`에서 플레이어를 Bark 별칭으로 라우팅해야 합니다**. 그렇지 않으면 미설정 플레이어는 기본적으로 Telegram으로 푸시되는데, Telegram 설정이 없으면 경고 한 줄만 출력하고 건너뛰므로 결국 아무것도 푸시되지 않습니다.

### 3. 경로 A: Telegram Bot 푸시만 사용(가장 간단)

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

3. 일상 사용: 업로드 서버를 시작합니다. 패킷 캡처 클라이언트(MitM 모듈 / Reqable 보고서 서버)가「업로드 API」에 따라 분할 업로드하면 지도가 자동으로 생성되고 푸시됩니다:

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

경로 A에서는 **필요하지 않습니다**: `config/push_map.json`, `config/bark_map.json`, 정적 파일 서버, `BARK_IMAGE_BASE`. 미설정 플레이어는 기본적으로 Telegram으로 푸시됩니다.

### 4. 경로 B: Bark 푸시 활성화(추가 설정 필요)

경로 A를 기반으로(Telegram 설정은 유지해도 되고, 비워 두고 Bark만 푸시해도 됩니다) 순서대로 다음을 채워 넣습니다:

1. **Bark key 설정**: `config/bark_map.json`에서 각 별칭에 대해 기기 key를 설정합니다(템플릿은 같은 디렉터리의 `bark_map.example.json` 참조).
2. **플레이어 라우팅 설정**: `config/push_map.json`에서 플레이어 ID를 Bark 별칭으로 라우팅합니다. 예:

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ⚠️ **반드시 설정**: 미설정 플레이어는 기본적으로 Telegram으로 푸시됩니다. 이때 Telegram도 설정되어 있지 않으면 경고만 출력하고 건너뛰므로 결과적으로 아무것도 푸시되지 않습니다.
3. **정적 파일 서버 구축**: 프로젝트의 `data/` 디렉터리를 공개 네트워크에서 접근 가능한 HTTP(S) 서비스로 노출하고, `.env`에 `BARK_IMAGE_BASE=https://<도메인 또는 IP:포트>`를 설정합니다. 그렇지 않으면 Bark 알림에 지도 이미지가 포함되지 않습니다(자세한 내용은 아래「정적 파일 서버 구축」참조).
4. 검증과 일상 사용은 경로 A(2, 3단계)와 동일합니다.

## 업로드 API

클라이언트가 캡처한 mysekai 응답 본문을 분할해 `POST /uploadMySekai`로 POST합니다(수동으로 curl을 사용해 동일한 프로토콜로 디버깅해도 됩니다). header는 다음과 같습니다:

| Header | 설명 |
| --- | --- |
| `X-Upload-Id` | 업로드 작업 ID(영숫자와 `-` / `_`만 허용, 길이 1~64), 필수 |
| `X-Chunk-Index` | 분할 번호, 0부터 시작, 필수 |
| `X-Total-Chunks` | 전체 분할 수(1~10), 필수 |
| `X-Original-Url` | 플레이어 ID를 파싱하는 데 사용하는 클라이언트 원본 페이지 URL(예: `https://.../user/123456...`); **선택**, 없으면 플레이어 ID는 `unknown`으로 기록됨 |
| `X-Script-Version` | 클라이언트 스크립트 버전 번호; 서버는 이 헤더를 무시하므로 보내지 않아도 됨 |

요청 본문은 원본 바이너리 분할 데이터입니다(multipart 불필요).

제한:

- 단일 파일 총 크기 ≤1MB(`MAX_TOTAL_SIZE`)
- 개별 분할 ≤1MB(`MAX_CHUNK_SIZE`, 초과 시 413 반환)
- 전체 분할 수 ≤10(`MAX_CHUNKS`)

> 참고: 총 크기 상한은 1MB뿐이므로, **분할 크기는 1MB보다 훨씬 작아야 의미가 있습니다**(예: 256KB면 10개 분할로 1MB를 모두 보낼 수 있음). 클라이언트가 1MB 분할을 사용하면 1MB를 초과하는 모든 파일은 2번째 분할부터 413으로 거부되어 사실상 단일 분할 업로드로만 동작합니다.

응답:

| 상태 코드 | 의미 |
| --- | --- |
| `200` | 분할을 수신하고 `OK`를 반환; 마지막 분할이 도착하면 서버가 자동으로 저장 파일 병합 → 지도 생성 → `data/archive/by-id/<user_id>/<타임스탬프>/`에 보관 → 알림 푸시까지 완료하며, 전 과정에서 수동 개입이 필요 없습니다 |
| `400` | 매개변수 오류(upload id 형식 오류, 분할 번호 범위 초과, 전체 분할 수가 1~10 범위 밖) |
| `413` | 크기 제한 초과(개별 분할이 1MB 초과 또는 누적 총 크기가 1MB 초과) |

### curl 예시

저장 파일이 ≤1MB이면 단일 분할로 전송이 끝납니다(가장 일반적):

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

## 푸시 메커니즘

### 기본은 Telegram Bot으로

- `config/push_map.json`에 설정되지 않은 플레이어는 **모두 기본적으로 Telegram으로 푸시**됩니다. `push_map.json` 파일이 없어도 마찬가지로 기본값은 Telegram입니다.
- Telegram은 Bot API `sendMediaGroup`을 사용해 로컬 PNG 4장을 multipart로 직접 업로드하므로 **공개 직링크도, 정적 파일 서버 의존도 필요 없습니다**. `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID`가 없으면 경고만 출력하고 건너뛰며, Bark 채널에는 영향을 주지 않습니다.

### Bark 푸시는 공개 직링크에 의존

Bark(Day.app) 알림의 이미지는 **URL 직링크**입니다: `notify.py`가 이미지 주소를 `image=` 매개변수에 인코딩해 `api.day.app`으로 보내고, Bark 서버가 그 이미지를 다시 가져옵니다. 따라서 해당 URL은 반드시 **공개 네트워크에서 접근 가능(HTTPS 권장)**해야 합니다. 그렇지 않으면 Bark 알림에 이미지가 포함되지 않습니다.

4장의 지도 직링크는 `notify.py`가 다음 우선순위에 따라 조합합니다:

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| 시나리오 | base 값 | 이미지 직링크 형태 |
| --- | --- | --- |
| 서버 흐름(권장) | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<타임스탬프>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<타임스탬프>/site_{5..8}.png` |
| 수동 CLI 푸시 | `BARK_IMAGE_BASE` 또는 `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`(`data/latest/`를 `<base>/` 아래에 노출해야 함) |

> 참고: 서버 흐름은 `BARK_IMAGE_BASE`가 설정된 경우에만 보관 경로가 포함된 직링크를 조합합니다. `FALLBACK_IMAGE_BASE`만 설정했다면 서버 푸시 직링크도 마찬가지로 `<FALLBACK_IMAGE_BASE>/site_{5..8}.png`입니다.

## 정적 파일 서버 예시(선택)

목적: `data/archive/` 디렉터리를 공개 URL로 노출해 Bark 서버가 네 장의 지도를 가져올 수 있게 합니다.

**권장 방법**: 정적 서버의 루트 디렉터리를 프로젝트의 `data/`로 지정하고 `BARK_IMAGE_BASE=https://<도메인 또는 IP:포트>`를 설정하면 자동으로 매핑됩니다:

```
data/archive/by-id/<user_id>/<타임스탬프>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<타임스탬프>/site_5.png
```

자주 쓰는 예시:

Python 내장(가장 간단, 내부 네트워크/테스트에 적합):

```bash
python -m http.server 8000 --directory data
# 그런 다음 BARK_IMAGE_BASE=http://<서버 IP>:8000 설정
```

nginx:

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    # ... ssl 인증서 설정 ...
    root /path/to/MySekaiMapper/data;
}
```

Caddy(자동 HTTPS):

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

주의사항:

- 직링크 주소로 **`127.0.0.1` / `localhost`를 사용하지 마십시오**. Bark 서버가 해당 주소에 접근할 수 있어야 하므로, 일반적으로 공개 네트워크에서 접근 가능한 주소를 선택하고, 내부 네트워크 IP는 상호 통신이 확인된 경우에만 사용합니다.
- **Telegram만 사용한다면 정적 서버가 전혀 필요 없습니다**. 이 섹션은 건너뛰어도 됩니다.
- 수동 `cli.py notify`의 직링크에는 보관 경로가 없으므로 `data/latest/`를 `BARK_IMAGE_BASE` 아래에 따로 노출해야 합니다. 또는 `FALLBACK_IMAGE_BASE`를 출력 디렉터리로 지정할 수 있습니다(예: `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → 해당 서버가 `data/latest/`를 `/output` 아래에 마운트).

## 플레이어 푸시 라우팅(선택)

필요에 따라 `config/` 아래에 로컬 설정 파일을 만듭니다(형식은 같은 디렉터리의 `*.example.json` 참조, `.gitignore`에 의해 무시됨):

- `push_map.json` — 플레이어 ID → 푸시 방식: 값은 `"telegram"`, Bark 별칭, `"none"`(푸시 안 함)이며, `["alias", "telegram"]` 또는 `"alias+tg"` 같은 조합 표기도 지원합니다. **미설정 플레이어의 기본값은 `telegram`**입니다.

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark 별칭 → 기기 key:

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 자주 묻는 질문

- **Bark 알림에 이미지가 오지 않나요?** 직링크가 공개 네트워크에서 접근 가능한지 확인하십시오: 브라우저/모바일 네트워크에서 `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<타임스탬프>/site_5.png`를 직접 열면 이미지가 표시되어야 합니다. 내부 네트워크 주소, `127.0.0.1`, 또는 인증서가 비정상적인 HTTPS는 모두 이미지 가져오기 실패를 유발합니다.
- **아무것도 푸시되지 않나요?** `push_map.json`에서 해당 플레이어를 `"none"`으로 설정했는지, Bark만 설정한 사용자가 해당 플레이어에 Bark 별칭 설정을 빠뜨렸는지(미설정 플레이어는 기본적으로 Telegram으로 푸시됨), Telegram 채널에 token과 chat id가 설정되었는지, Bark 채널에 key가 빠졌는지(`[BARK] ... failed` 로그 발생) 확인하십시오.
- **Bark는 싫고 Telegram만 원하나요?** 아무것도 할 필요가 없습니다——미설정 플레이어는 기본적으로 Telegram으로 푸시됩니다.

## 명령줄 도구(cli.py)

모든 기능은 `cli.py`로 실행할 수 있습니다. 설치 후(`pip install -e .`)에는 동등한 `mysekai` 명령도 사용할 수 있습니다. 명령이 성공하면 종료 코드는 0, 오류가 나면 1입니다(오류 메시지는 stderr로 출력).

```bash
python cli.py --help           # 하위 명령 개요
python cli.py <명령> --help     # 특정 하위 명령의 매개변수 확인
```

### generate —— 저장 파일 복호화 및 지도 생성

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`: 암호화된 저장 파일 경로(.bin), 필수
- 흐름: AES-128-CBC 복호화 → msgpack 파싱 → 드롭 좌표 추출 → 4장의 지도 생성(`site_5.png` ~ `site_8.png`) → `rare_resources.txt` 작성
- 출력은 `data/latest/`로 이루어지며, 종료 시 실제 경로를 출력합니다
- 사전 요구사항: `.env`에 `AES_KEY` / `AES_IV`가 설정되어 있어야 합니다. 저장 파일에 드롭 지점이 하나도 없으면 오류를 내고 종료합니다

### notify —— 지도와 통계 푸시

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`: `site_*.png`와 `rare_resources.txt`를 포함한 디렉터리(보통 `data/latest/`)
- `[task_id]`: 선택, 업로드 작업 ID, 기본값 `unknown`. `data/raw_mysekai/`에서 플레이어 ID를 역추적하는 데 사용: `mysekai_<플레이어ID>_<task_id>.bin`을 우선 매칭하고, 매칭되지 않으면 raw_mysekai의 최신 저장 파일을 사용
- Telegram으로 푸시할지 Bark로 푸시할지는 `config/push_map.json` 라우팅이 결정합니다(미설정 플레이어는 기본적으로 Telegram), 자세한 내용은「플레이어 푸시 라우팅」참조

### server —— 분할 업로드 서버 시작

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- FastAPI 서비스를 시작하며, 클라이언트가 `POST /uploadMySekai`로 암호화된 저장 파일을 분할 업로드합니다(API 세부 사항은「업로드 API」참조)
- 모든 분할이 도착하면 자동으로 저장 파일 병합 → 지도 생성 → `data/archive/by-id/<user_id>/<타임스탬프>/`에 보관 → 플레이어 라우팅에 따라 알림 푸시까지 완료하며, 수동 개입이 필요 없습니다
- 기본적으로 `9478` 포트를 수신합니다. 공개 네트워크에 배포할 때는 리버스 프록시로 HTTPS로 노출하는 것을 권장하며, 클라이언트 스크립트에 하드코딩된 업로드 URL(포트 포함)은 실제 배포 환경과 일치해야 합니다

### 대표적인 수동 흐름

```bash
python cli.py generate mysekai_xxx.bin       # 1. data/latest/에 지도 생성
python cli.py notify data/latest <task_id>   # 2. 푸시(task_id에 업로드 ID 입력, 예: chfto53c3)
```

## 디렉터리 구조

```
├── app/                       # 핵심 패키지
│   ├── config.py              # 경로／환경 변수／로컬 설정 통합 관리
│   ├── crypto.py              # MySekai 저장 파일 AES-128-CBC 복호화
│   ├── parser.py              # msgpack 파싱＋사이트 좌표 회전(순수 함수)
│   ├── render.py              # 드롭 지점 추출 → matplotlib 플로팅＋희귀 자원 통계
│   ├── notify.py              # 푸시: Telegram 미디어 그룹／Bark, 플레이어별 라우팅
│   ├── server.py              # FastAPI 분할 업로드 서비스
│   └── cli.py                 # 명령줄 진입점
├── assets/                    # 정적 리소스(저장소에 커밋됨)
│   ├── resourceId.csv         # 아이템 ID → 이름＋아이콘(base64)
│   └── NotoSansSC-Regular.ttf # 중국어 폰트(OFL 라이선스)
├── config/                    # 로컬 설정(실제 파일은 커밋하지 않음, *.example.json 참조)
│   ├── bark_map.example.json  # Bark 별칭 → 기기 key 템플릿
│   └── push_map.example.json  # 플레이어 ID → 푸시 방식 템플릿
├── data/                      # 런타임 데이터(디렉터리 전체 gitignore)
│   ├── tmp/                   # 분할 업로드 임시 저장, 병합 후 즉시 삭제
│   ├── raw_mysekai/           # 병합된 원본(암호화) 저장 파일, 영구 보존
│   ├── archive/               # 완성품 보관 by-id/<user>/<타임스탬프>/(Bark 직링크가 이곳을 가리킴)
│   └── latest/                # 가장 최근에 생성된 완성품
├── cli.py                     # 통합 진입점
├── tests/                     # 단위 테스트(pytest)
├── .env.example               # 환경 변수 템플릿(복사하여 .env로 작성)
└── requirements.txt           # 런타임 의존성(정확한 버전 고정)
```

## 테스트

```bash
python -m pytest
```

## 면책 조항

본 도구는 개인 학습과 오락 목적으로만 사용하시기 바라며, 상업적 용도나 게임 서비스 약관을 위반하는 행위에는 사용하지 마십시오. 게임 데이터와 아트 리소스의 저작권은 원저작권자에게 있습니다.

## 라이선스

이 프로젝트의 코드는 [MIT License](LICENSE)(저작권 © 2025 mouse233)를 따르며, 자유롭게 사용·수정·재배포할 수 있습니다. 자세한 내용은 [LICENSE](LICENSE)를 참조하십시오.

> ⚠️ 라이선스는 이 프로젝트의 코드에만 적용됩니다: `assets/`의 게임 자료(예: `resourceId.csv`의 아이템 아이콘)와 게임 데이터의 저작권은 SEGA / Colorful Palette 등 원저작권자에게 있으며, **MIT 라이선스 범위에 포함되지 않습니다**. 본 도구 이외의 용도로 사용하지 마십시오.
