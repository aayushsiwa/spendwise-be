package secure

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
)

var (
	encryptionKey []byte
	ErrKeyNotSet  = errors.New("encryption key not set")
	ErrInvalidKey = errors.New("invalid encryption key")
	newCipher     = aes.NewCipher
)

// SetKey stores the encryption key after validating that it is exactly 32 bytes. It returns ErrInvalidKey if the key length is invalid, nil otherwise.
func SetKey(key []byte) error {
	if len(key) != 32 {
		slog.ErrorContext(context.Background(), "Invalid encryption key length", "expected", 32, "got", len(key))
		return ErrInvalidKey
	}
	encryptionKey = key
	slog.InfoContext(context.Background(), "Encryption key set successfully")
	return nil
}

// GetKey returns the current encryption key (for testing purposes)
func GetKey() []byte {
	return encryptionKey
}

// Encrypt encrypts a plaintext string using AES-GCM, returning a base64-encoded result.
// It returns ErrKeyNotSet if no key has been set via SetKey. If the plaintext is empty, it returns an empty string without error.
func Encrypt(plain string) (string, error) {
	if len(encryptionKey) == 0 {
		slog.ErrorContext(context.Background(), "Encryption attempted without key")
		return "", ErrKeyNotSet
	}

	if plain == "" {
		return "", nil
	}

	block, err := newCipher(encryptionKey)
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to create cipher", "error", err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to create GCM", "error", err)
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		slog.ErrorContext(context.Background(), "Failed to generate nonce", "error", err)
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plain), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	slog.DebugContext(context.Background(), "Text encrypted successfully", "length", len(plain))
	return encoded, nil
}

// Decrypt restores plaintext from a base64-encoded AES-GCM encrypted string.
// It returns an error if the encryption key is not set, base64 decoding fails,
// or AES-GCM decryption fails. An empty input string returns an empty string and nil error.
func Decrypt(encoded string) (string, error) {
	if len(encryptionKey) == 0 {
		slog.ErrorContext(context.Background(), "Decryption attempted without key")
		return "", ErrKeyNotSet
	}

	if encoded == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to decode base64", "error", err)
		return "", err
	}

	block, err := newCipher(encryptionKey)
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to create cipher for decryption", "error", err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to create GCM for decryption", "error", err)
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		slog.ErrorContext(context.Background(), "Ciphertext too short", "length", len(data), "expected_min", nonceSize)
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to decrypt", "error", err)
		return "", err
	}

	slog.DebugContext(context.Background(), "Text decrypted successfully", "length", len(plain))
	return string(plain), nil
}

// SafeEncrypt encrypts text, returning an empty string if encryption fails.
func SafeEncrypt(plain string) string {
	encrypted, err := Encrypt(plain)
	if err != nil {
		slog.WarnContext(context.Background(), "Encryption failed, returning empty string", "error", err)
		return ""
	}
	return encrypted
}

// SafeDecrypt decrypts encoded text. It returns the decrypted string, or an empty string if decryption fails.
func SafeDecrypt(encoded string) string {
	decrypted, err := Decrypt(encoded)
	if err != nil {
		slog.WarnContext(context.Background(), "Decryption failed, returning empty string", "error", err)
		return ""
	}
	return decrypted
}
