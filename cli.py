#!/usr/bin/env python3
"""统一入口:`python cli.py generate|notify|server ...`。"""
from app.cli import main

if __name__ == "__main__":
    raise SystemExit(main())
