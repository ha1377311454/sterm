package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const encryptedPasswordPrefix = "enc:v1:"

func encryptPassword(keyFile, password string) (string, error) {
	if password == "" || strings.HasPrefix(password, encryptedPasswordPrefix) {
		return password, nil
	}
	key, err := loadOrCreateKey(keyFile)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(password), nil)
	return encryptedPasswordPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func decryptPassword(keyFile, password string) (string, error) {
	if password == "" || !strings.HasPrefix(password, encryptedPasswordPrefix) {
		return password, nil
	}
	key, err := loadOrCreateKey(keyFile)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(password, encryptedPasswordPrefix))
	if err != nil {
		return "", fmt.Errorf("decode encrypted password: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("encrypted password payload is too short")
	}
	nonce, cipherText := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %w", err)
	}
	return string(plain), nil
}

func loadOrCreateKey(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		path = filepath.Join(defaultConfigDir(), "key")
	}
	if data, err := os.ReadFile(path); err == nil {
		keyText := strings.TrimSpace(string(data))
		key, decErr := base64.StdEncoding.DecodeString(keyText)
		if decErr == nil && len(key) == 32 {
			return key, nil
		}
		if len(data) == 32 {
			return data, nil
		}
		return nil, fmt.Errorf("invalid AES key file %q: expected base64-encoded 32-byte key", path)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	keyText := base64.StdEncoding.EncodeToString(key) + "\n"
	if err := os.WriteFile(path, []byte(keyText), 0o600); err != nil {
		return nil, err
	}
	return key, nil
}
