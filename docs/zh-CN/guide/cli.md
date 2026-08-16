# 命令行工具（cli.py）

所有功能都可通过 `cli.py` 驱动；安装后（`pip install -e .`）也可用等价的 `mysekai` 命令。命令成功退出码为 0，出错为 1（错误信息打印到 stderr）。

```bash
python cli.py --help           # 子命令总览
python cli.py <命令> --help     # 查看某子命令的参数
```

## generate —— 解密存档并生成地图

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`：加密存档路径（.bin），必填
- 流程：AES-128-CBC 解密 → msgpack 解析 → 提取掉落坐标 → 绘制 4 张地图（`site_5.png` ~ `site_8.png`）→ 写出 `rare_resources.txt`
- 输出到 `data/latest/`，结束时打印实际路径
- 前置要求：`.env` 已配置 `AES_KEY` / `AES_IV`；存档中没有任何掉落点时会报错退出

## notify —— 推送地图与统计

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`：包含 `site_*.png` 与 `rare_resources.txt` 的目录（通常就是 `data/latest/`）
- `[task_id]`：可选，上传任务 ID，默认 `unknown`。用于从 `data/raw_mysekai/` 反查玩家 ID：优先匹配 `mysekai_<玩家ID>_<task_id>.bin`，匹配不到时取 raw_mysekai 里最新的存档
- 推送到 Telegram 还是 Bark 由 `config/push_map.json` 路由（未配置的玩家默认走 Telegram），详见[玩家推送路由](/zh-CN/guide/routing)

## server —— 启动分片上传服务

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- 启动 FastAPI 服务，客户端向 `POST /uploadMySekai` 分片上传加密存档（接口细节见[上传接口](/zh-CN/guide/upload-api)）
- 全部片到达后自动完成：合并存档 → 生成地图 → 归档到 `data/archive/by-id/<user_id>/<时间戳>/` → 按玩家路由推送通知，无需人工介入
- 默认监听 `9478` 端口；公网部署时建议通过反向代理暴露为 HTTPS，客户端脚本中写死的上传 URL（含端口）需与你的实际部署保持一致

## 典型手动流程

```bash
python cli.py generate mysekai_xxx.bin       # 1. 生成地图到 data/latest/
python cli.py notify data/latest <task_id>   # 2. 推送（task_id 填上传 ID，如 chfto53c3）
```
