<!-- GENERATED from README.md; do not edit directly. -->

# MySekaiMapper

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/>

🌐 **Languages**: [English](./) · [简体中文](./zh-CN/) · [繁體中文](./zh-TW/) · [日本語](./ja-JP/) · [한국어](./ko-KR/)

A Go service that turns encrypted *Project SEKAI* MySekai saves into resource-gathering maps and sends the result to Telegram or Bark (Day.app).

It works with a MitM capture client or Reqable's **Report Server**: the capture tool uploads a MySekai save, the service decrypts and parses it, renders maps and a rare-resource summary, archives the artifacts, and dispatches notifications without a manual processing step.

The usual MySekai areas produce `site_5.png` (Grassland), `site_6.png` (Beach), `site_7.png` (Flower Garden), `site_8.png` (Memorial Place), and `rare_resources.txt`. The renderer and notifier also handle any additional regular `site_*.png` outputs.

The capture flow has been verified on the CN and TW servers operated by Nuverse. Availability on other regions depends on their API path and save format.
