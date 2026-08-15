"""notify 模块测试:推送方式解析与 Bark key 选择。"""
import json

from app import config
from app.notify import get_bark_key_for, resolve_method


def test_resolve_method_legacy_strings():
    assert resolve_method(None) == ([], False)
    assert resolve_method("none") == ([], False)
    assert resolve_method("telegram") == ([], True)
    assert resolve_method("klee+tg") == (["klee"], True)
    assert resolve_method("dodoco") == (["dodoco"], False)


def test_resolve_method_list():
    assert resolve_method(["telegram", "dodoco"]) == (["dodoco"], True)
    assert resolve_method([]) == ([], False)


def test_bark_key_precedence(monkeypatch, tmp_path):
    bark_map = tmp_path / "bark_map.json"
    bark_map.write_text(json.dumps({"klee": "alias-key"}), encoding="utf-8")
    monkeypatch.setattr(config, "BARK_MAP_FILE", bark_map)

    # 1. explicit_key 优先
    assert get_bark_key_for(explicit_key="explicit") == "explicit"
    # 2. alias 查表
    assert get_bark_key_for(alias="klee") == "alias-key"
    # 3. alias 未命中 -> None
    assert get_bark_key_for(alias="nope") is None
    # 4. 无任何信息 -> None
    assert get_bark_key_for() is None
