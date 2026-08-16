# 정적 파일 서버 예시(선택)

목적: `data/archive/` 디렉터리를 공개 URL로 노출해 Bark 서버가 네 장의 지도를 가져올 수 있게 합니다.

**권장 방법**: 정적 서버의 루트 디렉터리를 프로젝트의 `data/`로 지정하고 `BARK_IMAGE_BASE=https://<도메인 또는 IP:포트>`를 설정하면 자동으로 매핑됩니다:

```
data/archive/by-id/<user_id>/<타임스탬프>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<타임스탬프>/site_5.png
```

## 자주 쓰는 예시

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

## 주의사항

- 직링크 주소로 **`127.0.0.1` / `localhost`를 사용하지 마십시오**. Bark 서버가 해당 주소에 접근할 수 있어야 하므로, 일반적으로 공개 네트워크에서 접근 가능한 주소를 선택하고, 내부 네트워크 IP는 상호 통신이 확인된 경우에만 사용합니다.
- **Telegram만 사용한다면 정적 서버가 전혀 필요 없습니다**. 이 섹션은 건너뛰어도 됩니다.
- 수동 `cli.py notify`의 직링크에는 보관 경로가 없으므로 `data/latest/`를 `BARK_IMAGE_BASE` 아래에 따로 노출해야 합니다. 또는 `FALLBACK_IMAGE_BASE`를 출력 디렉터리로 지정할 수 있습니다(예: `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → 해당 서버가 `data/latest/`를 `/output` 아래에 마운트).
