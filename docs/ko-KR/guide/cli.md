# 명령줄 도구(cli.py)

모든 기능은 `cli.py`로 실행할 수 있습니다. 설치 후(`pip install -e .`)에는 동등한 `mysekai` 명령도 사용할 수 있습니다. 명령이 성공하면 종료 코드는 0, 오류가 나면 1입니다(오류 메시지는 stderr로 출력).

```bash
python cli.py --help           # 하위 명령 개요
python cli.py <명령> --help     # 특정 하위 명령의 매개변수 확인
```

## generate —— 저장 파일 복호화 및 지도 생성

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`: 암호화된 저장 파일 경로(.bin), 필수
- 흐름: AES-128-CBC 복호화 → msgpack 파싱 → 드롭 좌표 추출 → 4장의 지도 생성(`site_5.png` ~ `site_8.png`) → `rare_resources.txt` 작성
- 출력은 `data/latest/`로 이루어지며, 종료 시 실제 경로를 출력합니다
- 사전 요구사항: `.env`에 `AES_KEY` / `AES_IV`가 설정되어 있어야 합니다. 저장 파일에 드롭 지점이 하나도 없으면 오류를 내고 종료합니다

## notify —— 지도와 통계 푸시

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`: `site_*.png`와 `rare_resources.txt`를 포함한 디렉터리(보통 `data/latest/`)
- `[task_id]`: 선택, 업로드 작업 ID, 기본값 `unknown`. `data/raw_mysekai/`에서 플레이어 ID를 역추적하는 데 사용: `mysekai_<플레이어ID>_<task_id>.bin`을 우선 매칭하고, 매칭되지 않으면 raw_mysekai의 최신 저장 파일을 사용
- Telegram으로 푸시할지 Bark로 푸시할지는 `config/push_map.json` 라우팅이 결정합니다(미설정 플레이어는 기본적으로 Telegram), 자세한 내용은 [플레이어 푸시 라우팅](/ko-KR/guide/routing) 참조

## server —— 업로드 서버 시작(분할 업로드 + Reqable 보고서 서버)

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- FastAPI 서비스를 시작합니다. 클라이언트는 `POST /uploadMySekai`로 암호화된 저장 파일을 업로드하고(단일 POST 또는 분할. API 세부 사항은 [업로드 API](/ko-KR/guide/upload-api) 참조), Reqable은 HAR 세션을 내장 보고서 엔드포인트로 보고할 수 있습니다([Reqable 보고서 서버](/ko-KR/guide/report-server) 참조)
- 모든 분할이 도착하면 자동으로 저장 파일 병합 → 지도 생성 → `data/archive/by-id/<user_id>/<타임스탬프>/`에 보관 → 플레이어 라우팅에 따라 알림 푸시까지 완료하며, 수동 개입이 필요 없습니다
- 기본적으로 `9478` 포트를 수신합니다. 공개 네트워크에 배포할 때는 리버스 프록시로 HTTPS로 노출하는 것을 권장하며, 클라이언트 스크립트에 하드코딩된 업로드 URL(포트 포함)은 실제 배포 환경과 일치해야 합니다

## 대표적인 수동 흐름

```bash
python cli.py generate mysekai_xxx.bin       # 1. data/latest/에 지도 생성
python cli.py notify data/latest <task_id>   # 2. 푸시(task_id에 업로드 ID 입력, 예: chfto53c3)
```
