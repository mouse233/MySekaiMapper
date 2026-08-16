---
layout: home

hero:
  name: MySekaiMapper
  text: MySekai 수집 포인트 지도 생성 도구
  tagline: 《프로젝트 세카이 컬러풀 스테이지! feat. 하츠네 미쿠》(Project Sekai) MySekai 수집 포인트 지도 생성 및 자동 푸시 도구입니다.
  actions:
    - theme: brand
      text: 빠른 시작
      link: /ko-KR/guide/introduction
    - theme: alt
      text: GitHub에서 보기
      link: https://github.com/mouse233/MySekaiMapper

features:
  - title: 완전 자동 파이프라인
    details: 패킷 캡처 도구(MitM 모듈 / Reqable 보고서 서버)가 MySekai 데이터 패킷을 분할 업로드하면 서버가 자동으로 병합·복호화·지도 생성·푸시까지 처리하며, 전 과정에서 수동 개입이 필요 없습니다.
  - title: 4장의 지도 + 희귀 자원 통계
    details: 작업마다 site_5.png ~ site_8.png(시작의 공터, 소원의 해변, 화려한 꽃밭, 잊혀진 곳)와 rare_resources.txt 희귀 자원 통계를 생성합니다.
  - title: Telegram 우선, Bark 지원
    details: Telegram은 multipart로 로컬 PNG 4장을 직접 전송하므로 공개 직링크가 필요 없습니다. 정적 파일 서버를 구성하면 Bark로도 이미지 직링크 알림을 받을 수 있습니다.
  - title: AES-128-CBC 복호화
    details: 암호화된 MySekai 저장 파일을 복호화하고 msgpack을 파싱해 사이트 좌표를 자동으로 회전한 뒤 matplotlib로 수집 지도를 그립니다.
---

## 작업 흐름

```
게임 API 응답 → MitM 모듈 / Reqable 보고서 서버(패킷 캡처로 mysekai 데이터 확보)
   │  ① 자동 분할 업로드 → server.py 자동 병합(권장, 프로젝트 취지)
   │  ② 또는 .bin 저장 파일 직접 배치 → cli.py generate
   ▼
parser.py    AES-128-CBC 복호화 + msgpack 파싱 + 좌표 회전
   ▼
render.py    site_5.png ~ site_8.png + rare_resources.txt 생성 → data/latest/
   ▼
notify.py    푸시:
             ├─ Telegram  : 이미지 multipart 직접 전송, 공개 직링크 불필요 ← 기본 채널
             └─ Bark      : image= URL 직링크 알림, 정적 파일 서버 필요
```

> ⚠️ **면책 조항**: 본 도구는 개인 학습과 오락 목적으로만 사용하시기 바라며, 상업적 용도나 게임 서비스 약관을 위반하는 행위에는 사용하지 마십시오. 게임 데이터와 아트 리소스의 저작권은 원저작권자에게 있습니다.

이 사이트는 [English](/) · [简体中文](/zh-CN/) · [繁體中文](/zh-TW/) · [日本語](/ja-JP/) · [한국어](/ko-KR/) 버전도 제공합니다.
