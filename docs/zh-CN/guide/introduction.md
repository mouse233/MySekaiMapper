<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 工作流程

```text
游戏 API 响应 → MitM 模块 / Reqable 上报服务器
    │  ① POST /uploadMySekai（单次或有序分片上传）
    │  ② POST /reqable/report（HAR，可选 gzip / br / zstd）
    ▼
mysekaimapper serve
    ├─ AES-128-CBC 解密 + MsgPack 解析 + 坐标归一化
    ├─ 绘制 site_*.png + rare_resources.txt
    ├─ 归档到 data/archive/by-id/<player_id>/<timestamp>/
    └─ 发布 data/latest/ 并通知
         ├─ Telegram：以 multipart 媒体组上传本地图片
         └─ Bark：从公开静态文件服务器读取图片 URL
```
