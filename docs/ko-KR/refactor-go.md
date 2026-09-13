<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# Go 리팩터링

현재 런타임은 Go만 사용합니다. 모듈은 `cmd/`, `internal/`, `go.mod`, `go.sum`으로 이루어진 표준 루트 구조를 따르며 Python 소스, 의존성 및 CI는 제거되었습니다. 보관된 참조 구현은 [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) 브랜치와 [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) 태그에 남아 있습니다.

HTTP 엔드포인트, 환경 변수, 출력 이름, 아카이브 레이아웃 및 라우팅 파일 형식은 호환성을 유지합니다. Go 렌더러는 고정 캔버스를 사용하므로 생성되는 PNG가 이전 Matplotlib 출력과 픽셀 단위로 동일하다고 보장되지는 않습니다.
