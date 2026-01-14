package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!2024",
		},
		{
			name:     "empty password",
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, tt.hash)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestJWTManager_GenerateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", time.Hour)
	userID := uuid.New()
	username := "testuser"

	token, err := jwtManager.GenerateToken(userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", time.Hour)
	userID := uuid.New()
	username := "testuser"

	token, err := jwtManager.GenerateToken(userID, username)
	require.NoError(t, err)

	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
}

func TestJWTManager_ValidateToken_InvalidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", time.Hour)

	tests := []struct {
		name        string
		token       string
		expectedErr error
	}{
		{
			name:        "invalid token format",
			token:       "invalid-token",
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "empty token",
			token:       "",
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtManager.ValidateToken(tt.token)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestJWTManager_ValidateToken_ExpiredToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", time.Nanosecond)
	userID := uuid.New()
	username := "testuser"

	token, err := jwtManager.GenerateToken(userID, username)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 10)

	_, err = jwtManager.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
