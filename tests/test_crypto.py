"""crypto 模块测试:AES 解密往返与密钥校验。"""
import pytest
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes

from app import crypto

KEY = b"0123456789abcdef"
IV = b"fedcba9876543210"


def _encrypt(plain: bytes) -> bytes:
    """用与游戏相同的 AES-128-CBC + 尾部填充方式加密。"""
    pad = 16 - len(plain) % 16
    data = plain + bytes([pad]) * pad
    cipher = Cipher(algorithms.AES(KEY), modes.CBC(IV))
    enc = cipher.encryptor()
    return enc.update(data) + enc.finalize()


def test_decrypt_roundtrip(monkeypatch):
    monkeypatch.setenv("AES_KEY", KEY.decode("ascii"))
    monkeypatch.setenv("AES_IV", IV.decode("ascii"))
    assert crypto.decrypt_mysekai(_encrypt(b"hello mysekai")) == b"hello mysekai"


def test_missing_key_raises(monkeypatch):
    monkeypatch.delenv("AES_KEY", raising=False)
    monkeypatch.delenv("AES_IV", raising=False)
    with pytest.raises(RuntimeError, match="AES_KEY"):
        crypto.get_aes_params()


def test_bad_key_length_raises(monkeypatch):
    monkeypatch.setenv("AES_KEY", "too-short")
    monkeypatch.setenv("AES_IV", IV.decode("ascii"))
    with pytest.raises(RuntimeError, match="16 bytes"):
        crypto.get_aes_params()
