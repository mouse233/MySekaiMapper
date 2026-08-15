"""解密存档 -> 提取掉落点 -> 绘制采集地图与稀有资源统计。"""
import base64
from io import BytesIO
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from PIL import Image
from matplotlib.font_manager import FontProperties
from matplotlib.offsetbox import OffsetImage, AnnotationBbox

from . import config
from .parser import decrypt_and_parse, extract_drops_from_obj, rotate_coords

ICON_ZOOM_MAP = 0.08
ICON_ZOOM_LEGEND = 0.085
VERT_OFFSET = 0.6
TEXT_DX = 0.6
FIGSIZE = (12, 12)
IGNORE_COUNT_IDS = {1, 6}
RARE_IDS = [5, 12, 20, 24]
SITE_NAME_MAP = {5: "初始空地", 6: "心愿沙滩", 7: "烂漫花田", 8: "忘却之所"}


# ================== 加载图标 ==================

def load_icons():
    icons_df = pd.read_csv(config.RESOURCE_CSV)
    icons_df["resourceId"] = pd.to_numeric(icons_df["resourceId"], errors="coerce").astype("Int64")

    name_col = None
    for c in ("name", "物品名", "itemName"):
        if c in icons_df.columns:
            name_col = c
            break

    icon_map = {}
    name_map = {}

    for _, row in icons_df.iterrows():
        rid = row["resourceId"]
        if pd.isna(rid):
            continue
        b64 = row.get("base64")
        if pd.isna(b64):
            continue
        try:
            img_bytes = base64.b64decode(str(b64))
            img = Image.open(BytesIO(img_bytes)).convert("RGBA")
            icon_map[int(rid)] = img
            name_map[int(rid)] = str(row[name_col]) if name_col else str(int(rid))
        except Exception:
            pass

    return icon_map, name_map


# ================== 生图 ==================

def draw_maps(drops: pd.DataFrame, icon_map, name_map):
    try:
        chinese_font = FontProperties(fname=config.FONT_FILE)
    except Exception:
        chinese_font = None

    sites = sorted(drops["mysekaiSiteId"].dropna().unique())

    for site in sites:
        plt.figure(figsize=(9, 9))
        ax = plt.gca()
        ax.set_aspect("equal", adjustable="box")
        plt.title(f"Harvest Drops — Site {int(site)}", fontsize=14)

        site_df = drops[drops["mysekaiSiteId"] == site]
        grouped = site_df.groupby(["positionX", "positionZ"])

        for (x, z), sub in grouped:
            xr, zr = rotate_coords(x, z, site)
            items = sorted(sub["resourceId"].dropna().unique())

            for i, rid in enumerate(items):
                rid = int(rid)
                if rid not in icon_map:
                    continue

                img_arr = np.array(icon_map[rid])
                imagebox = OffsetImage(img_arr, zoom=ICON_ZOOM_MAP)
                y_offset = -i * VERT_OFFSET
                ab = AnnotationBbox(imagebox, (xr, zr + y_offset), frameon=False, zorder=10)
                ax.add_artist(ab)

                count = int((sub["resourceId"] == rid).sum())
                if rid in IGNORE_COUNT_IDS:
                    continue
                if count > 1:
                    ax.text(
                        xr + TEXT_DX,
                        zr + y_offset,
                        f"×{count}",
                        fontsize=10,
                        fontproperties=chinese_font,
                        va="center",
                        ha="left",
                        zorder=11,
                    )

        rotated = [rotate_coords(x, z, site) for (x, z) in grouped.groups.keys()]
        if rotated:
            xs, zs = zip(*rotated)
            ax.set_xlim(min(xs) - 2, max(xs) + 2)
            ax.set_ylim(min(zs) - 2, max(zs) + 2)

        plt.xlabel("positionX")
        plt.ylabel("positionZ")

        out_path = config.LATEST_DIR / f"site_{int(site)}.png"
        plt.tight_layout()
        plt.savefig(out_path, dpi=200, bbox_inches="tight")
        plt.close()

        print(f"[OK] Generated {out_path}")


# ================== 稀有资源统计 ==================

def write_rare_resources_txt(drops: pd.DataFrame, name_map: dict):
    lines = ["稀有资源统计\n"]

    for rid in RARE_IDS:
        try:
            count = int((drops["resourceId"] == rid).sum())
        except Exception:
            count = 0

        try:
            sites = drops.loc[drops["resourceId"] == rid, "mysekaiSiteId"].dropna().unique().tolist()
            sites_clean = []
            for s in sites:
                try:
                    sites_clean.append(int(s))
                except Exception:
                    continue
            sites_clean = sorted(set(sites_clean))
        except Exception:
            sites_clean = []

        site_names = [SITE_NAME_MAP.get(s, f"site{s}") for s in sites_clean]
        site_suffix = f"（{'、'.join(site_names)}）" if site_names else ""

        name = name_map.get(rid, str(rid))
        lines.append(f"{name} × {count} {site_suffix}")

    out_path = config.LATEST_DIR / "rare_resources.txt"
    out_path.write_text("\n".join(lines), encoding="utf-8")
    print(f"[OK] Rare resource stats saved to {out_path}")


# ================== 主流程 ==================

def generate(mysekai_path) -> Path:
    """完整流程:解密 -> 提取 -> 绘图 -> 统计。

    返回输出目录(config.LATEST_DIR)。
    """
    mysekai_path = Path(mysekai_path)
    print("[*] Reading:", mysekai_path)

    obj = decrypt_and_parse(mysekai_path.read_bytes())

    print("[*] Extracting drops...")
    drops = extract_drops_from_obj(obj)
    if drops.empty:
        raise RuntimeError("No drops found in the save file")

    print("[*] Loading icons...")
    icon_map, name_map = load_icons()

    print("[*] Drawing maps...")
    draw_maps(drops, icon_map, name_map)

    print("[*] Writing rare resource stats...")
    write_rare_resources_txt(drops, name_map)

    print("[DONE] All maps generated")
    return config.LATEST_DIR
