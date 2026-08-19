<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# 서비스 실행

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

서버는 준비된 URL을 출력하고 업로드/보고서 수락, 큐 등록, 파싱, 렌더링, 보관, 알림, 경과 시간, 작업 ID 및 `player_id`에 대한 수명 주기 로그를 기록합니다. 보관 데이터 본문, 비밀값, 토큰 또는 전체 알림 URL은 의도적으로 로그에 남기지 않습니다.

프로세스는 `SIGINT` 및 `SIGTERM`을 처리합니다. HTTP 요청 수락을 중지한 뒤, 이미 수락된 작업을 최대 15초 동안 처리합니다.

컴파일된 바이너리는 `--root /path/to/MySekaiMapper`를 사용하여 체크아웃 외부에서 실행할 수 있습니다. 그렇지 않으면 작업 디렉터리를 기준으로 저장소 루트를 찾습니다.
