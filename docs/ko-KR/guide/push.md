# 푸시 메커니즘

## 기본은 Telegram Bot으로

- `config/push_map.json`에 설정되지 않은 플레이어는 **모두 기본적으로 Telegram으로 푸시**됩니다. `push_map.json` 파일이 없어도 마찬가지로 기본값은 Telegram입니다.
- Telegram은 Bot API `sendMediaGroup`을 사용해 로컬 PNG 4장을 multipart로 직접 업로드하므로 **공개 직링크도, 정적 파일 서버 의존도 필요 없습니다**. `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID`가 없으면 경고만 출력하고 건너뛰며, Bark 채널에는 영향을 주지 않습니다.

## Bark 푸시는 공개 직링크에 의존

Bark(Day.app) 알림의 이미지는 **URL 직링크**입니다: `notify.py`가 이미지 주소를 `image=` 매개변수에 인코딩해 `api.day.app`으로 보내고, Bark 서버가 그 이미지를 다시 가져옵니다. 따라서 해당 URL은 반드시 **공개 네트워크에서 접근 가능(HTTPS 권장)**해야 합니다. 그렇지 않으면 Bark 알림에 이미지가 포함되지 않습니다.

4장의 지도 직링크는 `notify.py`가 다음 우선순위에 따라 조합합니다:

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| 시나리오 | base 값 | 이미지 직링크 형태 |
| --- | --- | --- |
| 서버 흐름(권장) | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<타임스탬프>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<타임스탬프>/site_{5..8}.png` |
| 수동 CLI 푸시 | `BARK_IMAGE_BASE` 또는 `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png`(`data/latest/`를 `<base>/` 아래에 노출해야 함) |

::: tip
서버 흐름은 `BARK_IMAGE_BASE`가 설정된 경우에만 보관 경로가 포함된 직링크를 조합합니다. `FALLBACK_IMAGE_BASE`만 설정했다면 서버 푸시 직링크도 마찬가지로 `<FALLBACK_IMAGE_BASE>/site_{5..8}.png`입니다.
:::

이미지를 공개 네트워크에 노출하는 방법은 [정적 파일 서버](/ko-KR/guide/static-server)를, 플레이어별 Telegram / Bark 배정 방법은 [플레이어 푸시 라우팅](/ko-KR/guide/routing)을 참조하십시오.
