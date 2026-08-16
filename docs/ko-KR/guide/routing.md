# 플레이어 푸시 라우팅(선택)

필요에 따라 `config/` 아래에 로컬 설정 파일을 만듭니다(형식은 같은 디렉터리의 `*.example.json` 참조, `.gitignore`에 의해 무시됨):

- `push_map.json` — 플레이어 ID → 푸시 방식: 값은 `"telegram"`, Bark 별칭, `"none"`(푸시 안 함)이며, `["alias", "telegram"]` 또는 `"alias+tg"` 같은 조합 표기도 지원합니다. **미설정 플레이어의 기본값은 `telegram`**입니다.

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark 별칭 → 기기 key:

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## 구성 템플릿

| 파일 | 용도 | 템플릿 |
| --- | --- | --- |
| `config/push_map.json` | 플레이어 ID → 푸시 방식 라우팅 | `config/push_map.example.json` |
| `config/bark_map.json` | Bark 별칭 → 기기 key | `config/bark_map.example.json` |
