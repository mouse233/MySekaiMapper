# 디렉터리 구조

```
├── app/                       # 핵심 패키지
│   ├── config.py              # 경로／환경 변수／로컬 설정 통합 관리
│   ├── crypto.py              # MySekai 저장 파일 AES-128-CBC 복호화
│   ├── parser.py              # msgpack 파싱＋사이트 좌표 회전(순수 함수)
│   ├── har.py                 # Reqable 보고서 서버 HAR 파싱 및 압축 해제(순수 함수)
│   ├── render.py              # 드롭 지점 추출 → matplotlib 플로팅＋희귀 자원 통계
│   ├── notify.py              # 푸시: Telegram 미디어 그룹／Bark, 플레이어별 라우팅
│   ├── server.py              # FastAPI 업로드 서비스(분할 업로드 + Reqable 보고서 서버)
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
