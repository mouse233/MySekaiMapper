# 프로젝트 소개

**MySekaiMapper**는 프로젝트 세카이 컬러풀 스테이지! feat. 하츠네 미쿠(Project Sekai)의 MySekai(내 세카이) 수집 포인트 지도 생성 도구입니다.

**프로젝트 취지**: MitM 모듈 또는 Reqable의「보고서 서버」기능과 함께 사용합니다——패킷 캡처 도구가 게임 내 MySekai 데이터 패킷을 캡처하면 자동으로 이 서비스에 업로드합니다(1회 POST로 전송 가능, 분할 업로드도 지원). 서버가 암호화된 저장 파일을 복호화하여 각 사이트의 자원 드롭 좌표를 추출한 뒤 수집 지도를 그려 그 결과(희귀 자원 통계 포함)를 플레이어의 Telegram / Bark(iOS Day.app)로 푸시합니다. 전 과정에서 수동 개입이 필요 없습니다.

한 번의 작업으로 **4장의 지도**가 생성됩니다: `site_5.png`(시작의 공터), `site_6.png`(소원의 해변), `site_7.png`(화려한 꽃밭), `site_8.png`(잊혀진 곳), 그리고 `rare_resources.txt` 희귀 자원 통계 1개가 추가로 생성됩니다.

::: info 서버 호환성
본 프로젝트는 Nuverse(조석광년)가 운영하는 CN 서버 / TW 서버에서 테스트를 통과했습니다. 다른 서버에서의 동작은 확인되지 않았습니다.
:::

## 작업 흐름

```
게임 API 응답 → MitM 모듈 / Reqable 보고서 서버(패킷 캡처로 mysekai 데이터 확보)
   │  ① 자동 업로드(1회 POST, 분할도 지원) → server.py 자동 처리
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

## 환경 요구사항

- Python 3.10+
- 실행 의존성은 `requirements.txt`를 기준으로 합니다(정확한 버전 고정)

## 빠른 탐색

- [빠른 시작](/ko-KR/guide/quickstart) — 설치, `.env` 설정, 경로 A 또는 경로 B 선택
- [업로드 API](/ko-KR/guide/upload-api) — 패킷 캡처 클라이언트용 분할 업로드 인터페이스
- [푸시 메커니즘](/ko-KR/guide/push) — Telegram / Bark 알림 동작 방식
- [명령줄 도구](/ko-KR/guide/cli) — `cli.py generate` / `notify` / `server`
