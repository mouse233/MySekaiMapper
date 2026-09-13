<!-- GENERATED from doc/README.ko-KR.md; do not edit directly. -->

# MySekaiMapper

🌐 **Languages**: [English](../) · [简体中文](../zh-CN/) · [繁體中文](../zh-TW/) · [日本語](../ja-JP/) · [한국어](../ko-KR/)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/ko-KR/>

암호화된 *Project SEKAI* MySekai 저장 데이터를 자원 채집 지도으로 변환하고, 결과를 Telegram 또는 Bark(Day.app)로 전송하는 Go 서비스입니다.

MitM 캡처 클라이언트 또는 Reqable의 **Report Server**와 함께 사용할 수 있습니다. 캡처 도구가 MySekai 저장 데이터를 업로드하면, 서비스가 이를 복호화하고 파싱한 뒤 지도와 희귀 자원 요약을 렌더링하고 결과물을 보관하며, 수동 처리 없이 알림을 전송합니다.

일반적인 MySekai 지역에서는 `site_5.png`(초원), `site_6.png`(해변), `site_7.png`(꽃밭), `site_8.png`(기념 장소) 및 `rare_resources.txt`가 생성됩니다. 렌더러와 알림 전송기는 추가로 생성되는 일반 `site_*.png` 출력도 처리합니다.

캡처 흐름은 Nuverse가 운영하는 CN 및 TW 서버에서 검증되었습니다. 다른 지역에서의 사용 가능 여부는 해당 지역의 API 경로와 저장 데이터 형식에 따라 달라집니다.
