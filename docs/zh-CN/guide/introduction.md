# 项目介绍

**MySekaiMapper** 是《世界计划 多彩舞台》（Project Sekai）MySekai（我的世界）采集点地图生成工具。

**项目初衷**：搭配 MitM 模块或 Reqable 的「上报服务器」功能使用——抓包工具捕获游戏内 MySekai 数据包后，自动上传到本服务（一次 POST 即可，分片上传亦受支持）；服务端解密加密存档、提取各站点的资源掉落坐标，绘制采集地图，再把结果（含稀有资源统计）推送到玩家的 Telegram / Bark（iOS Day.app），全程无需人工介入。

一次任务会生成 **4 张地图**：`site_5.png`（初始空地）、`site_6.png`（心愿沙滩）、`site_7.png`（烂漫花田）、`site_8.png`（忘却之所），外加一份 `rare_resources.txt` 稀有资源统计。

::: info 服务器兼容性
本项目已在朝夕光年（Nuverse）运营的 CN 服 / TW 服中测试通过，其他服务器可用性未知。
:::

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

## 环境要求

- Python 3.10+
- 运行依赖以 `requirements.txt` 为准（精确锁版本）

## 快速导航

- [快速上手](/zh-CN/guide/quickstart) — 安装、配置 `.env`、选择路径 A 或路径 B
- [上传接口](/zh-CN/guide/upload-api) — 供抓包客户端使用的分片上传接口
- [推送机制](/zh-CN/guide/push) — Telegram / Bark 通知如何工作
- [命令行工具](/zh-CN/guide/cli) — `cli.py generate` / `notify` / `server`
