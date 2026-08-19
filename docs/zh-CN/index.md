<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# MySekaiMapper

🌐 **Languages**: [English](../) · [简体中文](../zh-CN/) · [繁體中文](../zh-TW/) · [日本語](../ja-JP/) · [한국어](../ko-KR/)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/zh-CN/>

MySekaiMapper 是面向 *Project SEKAI* MySekai 存档的 Go 服务：将加密存档转换为采集点地图，并将结果发送到 Telegram 或 Bark（Day.app）。

它可配合 MitM 抓包客户端或 Reqable 的 **上报服务器（Report Server）** 使用：抓包工具上传 MySekai 存档，服务解密、解析并绘制地图和稀有资源摘要，归档产物后自动发送通知，无需手动处理。

常见区域会生成 `site_5.png`（草地）、`site_6.png`（海滩）、`site_7.png`（花园）、`site_8.png`（纪念地）和 `rare_resources.txt`。渲染器与通知器也支持额外的常规 `site_*.png` 文件。

抓包流程已经在朝夕光年运营的国服和台服验证；其他地区是否可用取决于 API 路径和存档格式。
