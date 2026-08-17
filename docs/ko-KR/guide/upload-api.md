# 업로드 API

클라이언트가 캡처한 mysekai 응답 본문을 `POST /uploadMySekai`로 업로드합니다(1회 POST로 전송 가능. 분할 업로드는 호환성을 위해 유지). 수동으로 curl을 사용해 동일한 프로토콜로 디버깅해도 됩니다. header는 다음과 같습니다:

| Header | 설명 |
| --- | --- |
| `X-Upload-Id` | 업로드 작업 ID(영숫자와 `-` / `_`만 허용, 길이 1~64), 필수 |
| `X-Chunk-Index` | 분할 번호, 0부터 시작(단일 POST에서는 항상 0), 필수 |
| `X-Total-Chunks` | 전체 분할 수(1~10; 단일 POST에서는 1), 필수 |
| `X-Original-Url` | 플레이어 ID를 파싱하는 데 사용하는 클라이언트 원본 페이지 URL(예: `https://.../user/123456...`); **선택**, 없으면 플레이어 ID는 `unknown`으로 기록됨 |
| `X-Script-Version` | 클라이언트 스크립트 버전 번호; 서버는 이 헤더를 무시하므로 보내지 않아도 됨 |

요청 본문은 원본 바이너리 세이브 데이터입니다(multipart 불필요).

## 제한

- 단일 파일 총 크기 ≤1MB(`MAX_TOTAL_SIZE`)
- 개별 분할 ≤1MB(`MAX_CHUNK_SIZE`, 초과 시 413 반환)
- 전체 분할 수 ≤10(`MAX_CHUNKS`)

::: tip
현재 세이브는 약 200KB이므로 **1회 POST로 전송이 끝납니다**. 분할 업로드는 기존 캡처 클라이언트와의 호환성을 위해 유지되며, 사용할 경우 각 분할을 1MB보다 훨씬 작게(예: 256KB) 유지하면 10개 분할로 1MB 상한을 모두 보낼 수 있습니다.
:::

## 응답

| 상태 코드 | 의미 |
| --- | --- |
| `200` | 세이브를 수신하고 `OK`를 반환; 서버가 자동으로 저장 파일 병합(분할인 경우) → 지도 생성 → `data/archive/by-id/<user_id>/<타임스탬프>/`에 보관 → 알림 푸시까지 완료하며, 전 과정에서 수동 개입이 필요 없습니다 |
| `400` | 매개변수 오류(upload id 형식 오류, 분할 번호 범위 초과, 전체 분할 수가 1~10 범위 밖) |
| `413` | 크기 제한 초과(개별 분할이 1MB 초과 또는 누적 총 크기가 1MB 초과) |

## curl 예시

단일 POST(현재 세이브는 1회로 전송 완료):

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

분할 업로드(선택, 기존 클라이언트 호환용. 분할당 256KB, 10개 분할로 1MB 상한 전송):

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
total=$(( ($(wc -c < "$file") + 262143) / 262144 ))
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_

i=0
for c in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H "X-Original-Url: https://example.com/user/1234567890123456789" \
    --data-binary @"$c"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

`200 OK`가 반환되면 세이브가 수신된 것입니다. 이후 파이프라인(분할인 경우 병합 → 지도 생성 → 보관 → 푸시)은 자동으로 완료됩니다. `127.0.0.1:9478`을 실제 서비스 주소로 바꾸고, `X-Upload-Id`는 `^[a-zA-Z0-9_-]{1,64}$` 패턴과 일치해야 합니다(예: `openssl rand -hex 5`로 생성한 임의 문자열).
