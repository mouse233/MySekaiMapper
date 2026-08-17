"""命令行入口。

用法:
    python cli.py generate <mysekai_bin>      # 解密存档并生成地图
    python cli.py notify <output_dir> [task]  # 推送通知
    python cli.py server [--host H] [--port P] # 启动上传服务

安装后( pip install -e . )也可直接用 `mysekai` 命令。
"""
import argparse
import sys
from pathlib import Path


def cmd_generate(args):
    from .render import generate
    print(f"[*] Generating maps from {args.mysekai_bin}")
    out = generate(Path(args.mysekai_bin))
    print(f"[DONE] Maps written to {out}")


def cmd_notify(args):
    from .notify import notify
    notify(Path(args.output_dir), args.task_id)


def cmd_server(args):
    import uvicorn
    from .server import app, print_api_banner
    print_api_banner(args.host, args.port)
    uvicorn.run(app, host=args.host, port=args.port)


def main(argv=None):
    parser = argparse.ArgumentParser(
        prog="mysekai",
        description="MySekaiMapper:Project Sekai MySekai 采集点地图生成与推送工具",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    p_gen = sub.add_parser("generate", help="解密存档并生成采集地图与稀有资源统计")
    p_gen.add_argument("mysekai_bin", help="加密存档路径(.bin)")
    p_gen.set_defaults(func=cmd_generate)

    p_not = sub.add_parser("notify", help="推送地图与统计到 Telegram / Bark")
    p_not.add_argument("output_dir", help="包含 site_*.png 与 rare_resources.txt 的目录")
    p_not.add_argument("task_id", nargs="?", default="unknown")
    p_not.set_defaults(func=cmd_notify)

    p_srv = sub.add_parser("server", help="启动分片上传服务")
    p_srv.add_argument("--host", default="0.0.0.0")
    p_srv.add_argument("--port", type=int, default=9478)
    p_srv.set_defaults(func=cmd_server)

    args = parser.parse_args(argv)
    try:
        return args.func(args) or 0
    except Exception as e:
        print(f"[ERROR] {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
