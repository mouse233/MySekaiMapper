"""MySekai 存档的 AES-128-CBC 解密。"""
import os

from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes

from . import config  # noqa: F401  确保 .env 已加载


def get_aes_params():
    """从环境读取并校验 AES 密钥,返回 (key, iv) 字节。

    .env 已在 app.config 导入时合并进 os.environ,因此这里按需读取,
    便于测试用 monkeypatch 注入任意密钥。
    """
    key = os.environ.get("AES_KEY")
    iv = os.environ.get("AES_IV")
    if not key or not iv:
        raise RuntimeError(
            "AES_KEY / AES_IV not set. Copy .env.example to .env and fill in the keys."
        )
    key_b, iv_b = key.encode("utf-8"), iv.encode("utf-8")
    if len(key_b) != 16 or len(iv_b) != 16:
        raise RuntimeError("AES_KEY and AES_IV must each be exactly 16 bytes (AES-128-CBC).")
    return key_b, iv_b


def decrypt_mysekai(data: bytes) -> bytes:
    """解密 MySekai 存档,返回 msgpack 明文。"""
    key, iv = get_aes_params()
    cipher = Cipher(algorithms.AES(key), modes.CBC(iv))
    decryptor = cipher.decryptor()
    padded_plain = decryptor.update(data) + decryptor.finalize()
    pad_len = padded_plain[-1]
    return padded_plain[:-pad_len]
