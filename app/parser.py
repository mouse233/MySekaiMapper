"""存档解析与坐标变换(纯函数,便于单测)。"""
import msgpack
import pandas as pd

from . import crypto


def decrypt_and_parse(raw: bytes):
    """解密并解析存档,返回 msgpack 对象。"""
    plain = crypto.decrypt_mysekai(raw)
    return msgpack.unpackb(plain, raw=False)


def extract_drops_from_obj(obj) -> pd.DataFrame:
    """从解析后的对象中提取所有采集掉落点。"""
    updated = obj.get("updatedResources", {})
    maps = updated.get("userMysekaiHarvestMaps", [])
    drops = []
    for site in maps:
        site_id = site.get("mysekaiSiteId")
        for d in site.get("userMysekaiSiteHarvestResourceDrops", []):
            drops.append({
                "mysekaiSiteId": site_id,
                "resourceId": d.get("resourceId"),
                "positionX": d.get("positionX"),
                "positionZ": d.get("positionZ"),
            })
    return pd.DataFrame(drops)


def rotate_coords(x, z, site_id):
    """把各站点原始坐标旋转到统一朝向。

    - site 6(心愿沙滩):顺时针 90° -> (x, z) -> (z, -x)
    - site 5/8(初始空地/忘却之所):逆时针 90° -> (x, z) -> (-z, x)
    - site 7(烂漫花田):旋转 180° -> (x, z) -> (-x, -z)
    """
    try:
        x = float(x)
        z = float(z)
    except Exception:
        return None, None
    if site_id == 6:
        return z, -x
    if site_id in (5, 8):
        return -z, x
    if site_id == 7:
        return -x, -z
    return x, z
