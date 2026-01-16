package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		expectErr bool
	}{
		{
			name:      "valid AES-128 key",
			key:       "1234567890123456", // 16 bytes
			expectErr: false,
		},
		{
			name:      "valid AES-192 key",
			key:       "123456789012345678901234", // 24 bytes
			expectErr: false,
		},
		{
			name:      "valid AES-256 key",
			key:       "12345678901234567890123456789012", // 32 bytes
			expectErr: false,
		},
		{
			name:      "invalid key size",
			key:       "123", // 3 bytes
			expectErr: true,
		},
		{
			name:      "empty key",
			key:       "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := NewEncryptor(tt.key)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, encryptor)
				assert.ErrorIs(t, err, ErrInvalidKeySize)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encryptor)
			}
		})
	}
}

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	encryptor, err := NewEncryptor("12345678901234567890123456789012") // 32 bytes for AES-256
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "Hello, World!",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "long text",
			plaintext: "This is a very long text that should be encrypted and decrypted correctly. It contains multiple sentences and should test the encryption algorithm properly.",
		},
		{
			name:      "special characters",
			plaintext: "Тест с русскими символами и спецсимволами: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "json data",
			plaintext: `{"username": "test", "password": "secret123", "email": "test@example.com"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := encryptor.Encrypt(tt.plaintext)
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)
			assert.NotEqual(t, tt.plaintext, ciphertext)

			decrypted, err := encryptor.Decrypt(ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptor_EncryptBytesDecryptBytes(t *testing.T) {
	encryptor, err := NewEncryptor("12345678901234567890123456789012") // 32 bytes for AES-256
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple bytes",
			data: []byte("Hello, World!"),
		},
		{
			name: "empty bytes",
			data: []byte(nil),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "large binary data",
			data: make([]byte, 1024), // 1KB of zeros
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := encryptor.EncryptBytes(tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)

			decrypted, err := encryptor.DecryptBytes(ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tt.data, decrypted)
		})
	}
}

func TestEncryptor_Decrypt_InvalidData(t *testing.T) {
	encryptor, err := NewEncryptor("12345678901234567890123456789012")
	require.NoError(t, err)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "invalid base64",
			ciphertext: "invalid-base64-data!!!",
		},
		{
			name:       "empty string",
			ciphertext: "",
		},
		{
			name:       "too short data",
			ciphertext: "YWJj", // base64 for "abc" - too short for nonce
		},
		{
			name:       "corrupted data",
			ciphertext: "dGVzdGRhdGF0aGF0aXNub3R2YWxpZA==", // valid base64 but not encrypted data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			assert.Error(t, err)
		})
	}
}

func TestEncryptor_DecryptBytes_InvalidData(t *testing.T) {
	encryptor, err := NewEncryptor("12345678901234567890123456789012")
	require.NoError(t, err)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "invalid base64",
			ciphertext: "invalid-base64-data!!!",
		},
		{
			name:       "empty string",
			ciphertext: "",
		},
		{
			name:       "too short data",
			ciphertext: "YWJj", // base64 for "abc" - too short for nonce
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.DecryptBytes(tt.ciphertext)
			assert.Error(t, err)
		})
	}
}

func TestEncryptor_DifferentKeys(t *testing.T) {
	key1 := "12345678901234567890123456789012"
	key2 := "abcdefghijklmnopqrstuvwxyz123456"

	encryptor1, err := NewEncryptor(key1)
	require.NoError(t, err)

	encryptor2, err := NewEncryptor(key2)
	require.NoError(t, err)

	plaintext := "Secret message"

	ciphertext, err := encryptor1.Encrypt(plaintext)
	require.NoError(t, err)

	_, err = encryptor2.Decrypt(ciphertext)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCiphertext)
}
