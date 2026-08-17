package mapper

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"os"
)

// ErrAESNotConfigured identifies an absent or malformed local AES setup
// without exposing key material in an error message.
var ErrAESNotConfigured = errors.New("AES parameters are not configured")

// AESParams reads the MySekai AES-128-CBC key material from the environment.
// Values are deliberately not returned in error messages.
func AESParams() ([]byte, []byte, error) {
	key := []byte(os.Getenv("AES_KEY"))
	iv := []byte(os.Getenv("AES_IV"))
	if len(key) == 0 || len(iv) == 0 {
		return nil, nil, fmt.Errorf("%w: AES_KEY / AES_IV not set", ErrAESNotConfigured)
	}
	if len(key) != aes.BlockSize || len(iv) != aes.BlockSize {
		return nil, nil, fmt.Errorf("%w: AES_KEY and AES_IV must each be exactly 16 bytes", ErrAESNotConfigured)
	}
	return key, iv, nil
}

// DecryptArchive decrypts the game save using AES-128-CBC and validates the
// PKCS#7 padding used by the Python implementation.
func DecryptArchive(raw []byte) ([]byte, error) {
	key, iv, err := AESParams()
	if err != nil {
		return nil, err
	}
	return decryptWithKeyIV(raw, key, iv)
}

func decryptWithKeyIV(raw, key, iv []byte) ([]byte, error) {
	if len(raw) == 0 || len(raw)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("encrypted archive length must be a non-zero multiple of %d", aes.BlockSize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("AES IV must be %d bytes", block.BlockSize())
	}

	plain := make([]byte, len(raw))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, raw)
	pad := int(plain[len(plain)-1])
	if pad < 1 || pad > block.BlockSize() || pad > len(plain) {
		return nil, fmt.Errorf("invalid PKCS#7 padding")
	}
	for _, b := range plain[len(plain)-pad:] {
		if int(b) != pad {
			return nil, fmt.Errorf("invalid PKCS#7 padding")
		}
	}
	return plain[:len(plain)-pad], nil
}
