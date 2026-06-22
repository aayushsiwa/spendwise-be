package secure

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"os"
	"testing"
)

var testKey = []byte("0123456789abcdef0123456789abcdef")
var testKeyAlt = []byte("abcdef0123456789abcdef0123456789")

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.DiscardHandler))
	os.Exit(m.Run())
}

func TestSetKey(t *testing.T) {
	_ = SetKey(testKey)

	t.Run("wrong length 0", func(t *testing.T) {
		if err := SetKey(nil); err == nil {
			t.Error("expected error for nil key")
		}
	})

	t.Run("wrong length 16", func(t *testing.T) {
		if err := SetKey([]byte("1234567890abcdef")); err == nil {
			t.Error("expected error for 16-byte key")
		}
	})

	t.Run("wrong length 33", func(t *testing.T) {
		if err := SetKey([]byte("1234567890abcdef1234567890abcdef1")); err == nil {
			t.Error("expected error for 33-byte key")
		}
	})

	t.Run("correct length 32", func(t *testing.T) {
		if err := SetKey(testKey); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestGetKey(t *testing.T) {
	key := []byte("abcdef0123456789abcdef0123456789")
	_ = SetKey(key)

	t.Run("returns the same key", func(t *testing.T) {
		got := GetKey()
		if !bytes.Equal(got, key) {
			t.Errorf("expected %v, got %v", key, got)
		}
	})
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	_ = SetKey(testKey)

	plaintexts := []string{"hello", "world", "a", "abc123!@#", "longer text with spaces and \nnewlines"}
	for _, pt := range plaintexts {
		t.Run("round_trip_"+pt, func(t *testing.T) {
			enc, err := Encrypt(pt)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}
			dec, err := Decrypt(enc)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}
			if dec != pt {
				t.Errorf("round trip mismatch: got %q, want %q", dec, pt)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	t.Run("empty string returns empty", func(t *testing.T) {
		_ = SetKey(testKey)
		result, err := Encrypt("")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("no key set returns error", func(t *testing.T) {
		encryptionKey = nil
		_, err := Encrypt("test")
		if err != ErrKeyNotSet {
			t.Errorf("expected ErrKeyNotSet, got %v", err)
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("empty string returns empty", func(t *testing.T) {
		_ = SetKey(testKey)
		result, err := Decrypt("")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("no key set returns error", func(t *testing.T) {
		encryptionKey = nil
		_, err := Decrypt("dGVzdA==")
		if err != ErrKeyNotSet {
			t.Errorf("expected ErrKeyNotSet, got %v", err)
		}
	})

	t.Run("bad base64 returns error", func(t *testing.T) {
		_ = SetKey(testKey)
		_, err := Decrypt("!!!not-base64!!!")
		if err == nil {
			t.Error("expected error for bad base64")
		}
	})

	t.Run("wrong key fails decrypt", func(t *testing.T) {
		_ = SetKey(testKey)
		enc, err := Encrypt("secret message")
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}
		_ = SetKey(testKeyAlt)
		_, err = Decrypt(enc)
		if err == nil {
			t.Error("expected error when decrypting with wrong key")
		}
	})
}

func TestSafeEncrypt(t *testing.T) {
	t.Run("returns empty on error", func(t *testing.T) {
		encryptionKey = nil
		result := SafeEncrypt("test")
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})

	t.Run("encrypts on success", func(t *testing.T) {
		_ = SetKey(testKey)
		result := SafeEncrypt("hello")
		if result == "" {
			t.Error("expected non-empty encrypted string")
		}
	})
}

func TestSafeDecrypt(t *testing.T) {
	t.Run("returns empty on error", func(t *testing.T) {
		encryptionKey = nil
		result := SafeDecrypt("test")
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})

	t.Run("decrypts on success", func(t *testing.T) {
		_ = SetKey(testKey)
		enc := SafeEncrypt("hello")
		result := SafeDecrypt(enc)
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

type mockBlock struct{}

func (mockBlock) BlockSize() int          { return 8 } // Block size of 8 will fail GCM
func (mockBlock) Encrypt(dst, src []byte) {}
func (mockBlock) Decrypt(dst, src []byte) {}

func TestEncrypt_ErrorPaths(t *testing.T) {
	t.Run("aes.NewCipher failure", func(t *testing.T) {
		// Bypassing SetKey check to set an invalid length key
		encryptionKey = []byte("too-short")
		defer func() { _ = SetKey(testKey) }()

		_, err := Encrypt("hello")
		if err == nil {
			t.Error("expected error due to invalid key length inside Encrypt")
		}
	})

	t.Run("rand.Reader failure", func(t *testing.T) {
		_ = SetKey(testKey)
		origReader := rand.Reader
		rand.Reader = errReader{}
		defer func() { rand.Reader = origReader }()

		_, err := Encrypt("hello")
		if err == nil {
			t.Error("expected error due to failing rand.Reader")
		}
	})

	t.Run("cipher.NewGCM failure", func(t *testing.T) {
		_ = SetKey(testKey)
		origNewCipher := newCipher
		newCipher = func(key []byte) (cipher.Block, error) {
			return mockBlock{}, nil
		}
		defer func() { newCipher = origNewCipher }()

		_, err := Encrypt("hello")
		if err == nil {
			t.Error("expected error due to invalid block size for GCM")
		}
	})
}

func TestDecrypt_ErrorPaths(t *testing.T) {
	t.Run("aes.NewCipher failure", func(t *testing.T) {
		// Bypassing SetKey check to set an invalid length key
		encryptionKey = []byte("too-short")
		defer func() { _ = SetKey(testKey) }()

		_, err := Decrypt("dGVzdA==")
		if err == nil {
			t.Error("expected error due to invalid key length inside Decrypt")
		}
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		_ = SetKey(testKey)
		// 11 bytes base64 encoded data, GCM nonce is 12 bytes
		shortData := base64.StdEncoding.EncodeToString(make([]byte, 11))
		_, err := Decrypt(shortData)
		if err == nil {
			t.Error("expected error for ciphertext shorter than nonce size")
		}
	})

	t.Run("cipher.NewGCM failure", func(t *testing.T) {
		_ = SetKey(testKey)
		origNewCipher := newCipher
		newCipher = func(key []byte) (cipher.Block, error) {
			return mockBlock{}, nil
		}
		defer func() { newCipher = origNewCipher }()

		_, err := Decrypt("dGVzdA==")
		if err == nil {
			t.Error("expected error due to invalid block size for GCM")
		}
	})
}
