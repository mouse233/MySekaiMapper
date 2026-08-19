<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# Reqable Report Server

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
