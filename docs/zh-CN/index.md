---
layout: home

hero:
  name: MySekaiMapper
  text: MySekai 采集点地图生成工具
  tagline: 《世界计划 多彩舞台》（Project Sekai）MySekai 采集点地图生成与自动推送工具。
  actions:
    - theme: brand
      text: 快速上手
      link: /zh-CN/guide/introduction
    - theme: alt
      text: 在 GitHub 查看
      link: https://github.com/mouse233/MySekaiMapper

features:
  - title: 全自动流水线
    details: 抓包工具（MitM 模块 / Reqable 上报服务器）上传 MySekai 数据包，服务端自动解密、绘图并推送，全程无需人工介入。
  - title: 4 张地图 + 稀有资源统计
    details: 每次任务生成 site_5.png ~ site_8.png（初始空地、心愿沙滩、烂漫花田、忘却之所），外加 rare_resources.txt 稀有资源统计。
  - title: Telegram 优先，Bark 就绪
    details: Telegram 通过 multipart 直传 4 张 PNG，无需公网直链；配置静态文件服务器后 Bark 即可收到图片直链通知。
  - title: AES-128-CBC 解密
    details: 解密加密的 MySekai 存档，解析 msgpack，自动旋转站点坐标，并用 matplotlib 绘制采集地图。
---

## 工作流程

```
游戏 API 响应 → MitM 模块 / Reqable 上报服务器（抓包捕获 mysekai 数据）
   │  ① 自动上传（一次 POST，分片亦支持）→ server.py 自动处理
   │  ② 或手动放置 .bin 存档 → cli.py generate
   ▼
parser.py    AES-128-CBC 解密 + msgpack 解析 + 坐标旋转
   ▼
render.py    绘制 site_5.png ~ site_8.png + rare_resources.txt → data/latest/
   ▼
notify.py    推送：
             ├─ Telegram  ：图片 multipart 直传，无需公网直链 ← 默认渠道
             └─ Bark      ：以 image= URL 直链通知，需静态文件服务器
```

> ⚠️ **免责声明**：本工具仅用于个人学习与娱乐，请勿用于任何商业用途或违反游戏服务条款的行为。游戏数据与美术资源版权归原版权方所有。
