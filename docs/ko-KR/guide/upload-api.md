<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 업로드 API

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
