"""parser 模块纯函数测试:坐标旋转与掉落点提取。"""
from app.parser import extract_drops_from_obj, rotate_coords


def test_rotate_identity_site():
    assert rotate_coords(1.0, 2.0, 3) == (1.0, 2.0)
    assert rotate_coords(1.0, 2.0, 9) == (1.0, 2.0)


def test_rotate_site6_clockwise_90():
    assert rotate_coords(1.0, 2.0, 6) == (2.0, -1.0)


def test_rotate_site5_counter_clockwise_90():
    assert rotate_coords(1.0, 2.0, 5) == (-2.0, 1.0)


def test_rotate_site8_same_as_site5():
    assert rotate_coords(1.0, 2.0, 8) == rotate_coords(1.0, 2.0, 5)


def test_rotate_site7_180():
    assert rotate_coords(1.0, 2.0, 7) == (-1.0, -2.0)


def test_rotate_bad_values_return_none():
    assert rotate_coords("abc", 2.0, 1) == (None, None)
    assert rotate_coords(None, None, 1) == (None, None)


def test_extract_drops_from_obj():
    obj = {"updatedResources": {"userMysekaiHarvestMaps": [
        {
            "mysekaiSiteId": 5,
            "userMysekaiSiteHarvestResourceDrops": [
                {"resourceId": 1, "positionX": 1.0, "positionZ": 2.0},
                {"resourceId": 5, "positionX": 3.0, "positionZ": 4.0},
            ],
        },
        {"mysekaiSiteId": 6, "userMysekaiSiteHarvestResourceDrops": []},
    ]}}

    df = extract_drops_from_obj(obj)
    assert len(df) == 2
    assert set(df["mysekaiSiteId"]) == {5}
    assert sorted(df["resourceId"].tolist()) == [1, 5]
    assert df.iloc[0]["positionX"] == 1.0


def test_extract_drops_empty():
    df = extract_drops_from_obj({})
    assert df.empty
