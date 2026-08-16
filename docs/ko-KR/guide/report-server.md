# Reqable 보고서 서버

Reqable의 내장「보고서 서버」기능(v2.20.0+)은 캡처한 각 HTTP 세션을 HAR JSON 형식으로 사용자 서버에 자동으로 POST합니다. gzip / brotli / zstd 압축을 선택할 수 있습니다. 보고서 엔드포인트는 **기본적으로 활성화**되어 있으며 분할 업로드와 공존합니다——`python cli.py server`로 두 엔드포인트가 모두 제공됩니다. 비활성화하려면 `REPORT_ENABLED=0`을 설정합니다:

```bash
python cli.py server
```

설정(`.env`):

| 변수 | 기본값 | 설명 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1`(활성화) | `0` / `false`로 설정하면 보고서 엔드포인트가 비활성화됩니다 |
| `REPORT_PATH` | `/reqable/report` | 엔드포인트 경로. Reqable의「업로드 경로」란에 이 값을 입력합니다 |
| `REPORT_MAX_SIZE` | `8` | HAR 본문 크기 상한(MB). 세이브 데이터 자체는 ≤1MB여야 하며, base64로 약 33% 커집니다 |
| `REPORT_TOKEN` | (비어 있음) | 선택적 공유 토큰. 설정하면 엔드포인트가 `X-Report-Token` 헤더를 요구합니다 |

## 처리 과정

보고서를 받을 때마다 서버는:

1. `Content-Encoding`(gzip / br / zstd)에 따라 본문을 압축 해제하고 HAR을 파싱합니다.
2. `log.entries`를 순회하며「응답 본문(폴백: 요청 본문)이 `AES_KEY` / `AES_IV`로 복호화되고 MySekai 세이브로 파싱되는」첫 번째 세션을 채택합니다. 규칙에 매칭되더라도 세이브와 무관한 트래픽은 건너뜁니다.
3. 세션 URL(`/user/<id>`)에서 플레이어 ID를 추출합니다.
4. 세이브를 `data/raw_mysekai/`에 저장하고, 분할 업로드와 동일한「생성 → 아카이브 → 푸시」파이프라인을 시작합니다.

::: warning
Reqable은 각 세션을 **정확히 1번만 전송하고 실패 시 재시도하지 않습니다**. 따라서 엔드포인트는 가능한 한 빨리 `200`을 반환합니다. 서버를 안정적으로 유지하고 `[REPORT]` 로그를 확인하세요.
:::

보고서 1건당 처리되는 세이브는 **1개뿐**(첫 번째 유효 항목)이므로, 여러 엔드포인트에 매칭되는 규칙이라도 중복 푸시가 발생하지 않습니다.

## 보안

프로토콜 자체에는 인증이 없습니다. Reqable은 사용자 정의 헤더를 추가할 수 없으므로, `REPORT_TOKEN`에 의존하기보다 `REPORT_PATH`에 임의 문자열을 포함시키거나(예: `/reqable/report/9f3a…`) 리버스 프록시 / 방화벽으로 접근을 제한하는 것이 좋습니다.

## Reqable 측 설정

- URL 매칭 규칙: `https://<게임API호스트>/*`(또는 `https://<게임API호스트>/user/*/mysekai*`처럼 좁힐 수 있음)
- 업로드 경로: `http://<내 서버>:9478/reqable/report`
- 압축 알고리즘: gzip / brotli / zstd 모두 가능(서버가 3가지 모두 지원)

5개 서버의 게임 API 호스트:

| 서버 | 게임 API 호스트 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

먼저 도메인 전체 규칙 `https://<도메인>/*`으로 시작하는 것을 권장합니다——무관한 세션은 서버가 자동으로 건너뜁니다. 사용 중인 서버의 mysekai API도 `/api/user/*/mysekai*` 경로(CN 실측 검증 완료)라면 `https://<도메인>/api/user/*/mysekai*`로 좁혀 업로드 양을 줄일 수 있습니다.

## curl 예시

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```
