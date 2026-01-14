package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Himany/GophKeeper/internal/config"
	"github.com/Himany/GophKeeper/pkg/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalStorage(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)
	assert.NotNil(t, storage)

	_, err = os.Stat(tempDir)
	assert.NoError(t, err)
}

func TestLocalStorage_SaveLoadToken(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	token := "test.jwt.token"

	err = storage.SaveToken(token)
	require.NoError(t, err)

	loadedToken, err := storage.LoadToken()
	require.NoError(t, err)
	assert.Equal(t, token, loadedToken)
}

func TestLocalStorage_LoadToken_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "nonexistent"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	token, err := storage.LoadToken()
	require.NoError(t, err)
	assert.Empty(t, token)
}

func TestLocalStorage_RemoveToken(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	token := "test.jwt.token"

	err = storage.SaveToken(token)
	require.NoError(t, err)

	err = storage.RemoveToken()
	require.NoError(t, err)

	loadedToken, err := storage.LoadToken()
	require.NoError(t, err)
	assert.Empty(t, loadedToken)
}

func TestLocalStorage_RemoveToken_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "nonexistent"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	err = storage.RemoveToken()
	assert.NoError(t, err)
}

func TestLocalStorage_SaveLoadCache(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	cacheData := &CacheData{
		Entries: []api.EntryResponse{
			{
				ID:   uuid.New(),
				Name: "Test Entry 1",
				Type: "credentials",
				Data: map[string]string{
					"login":    "user1",
					"password": "pass1",
				},
			},
			{
				ID:   uuid.New(),
				Name: "Test Entry 2",
				Type: "text",
				Data: map[string]string{
					"content": "Some text",
				},
			},
		},
		LastVersion: 42,
		UpdatedAt:   time.Now().Unix(),
	}

	err = storage.SaveCache(cacheData)
	require.NoError(t, err)

	loadedCache, err := storage.LoadCache()
	require.NoError(t, err)
	assert.NotNil(t, loadedCache)
	assert.Len(t, loadedCache.Entries, 2)
	assert.Equal(t, "Test Entry 1", loadedCache.Entries[0].Name)
	assert.Equal(t, "Test Entry 2", loadedCache.Entries[1].Name)
	assert.Equal(t, int64(42), loadedCache.LastVersion)
	assert.Equal(t, cacheData.UpdatedAt, loadedCache.UpdatedAt)
}

func TestLocalStorage_LoadCache_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	cache, err := storage.LoadCache()
	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Empty(t, cache.Entries)
	assert.Equal(t, int64(0), cache.LastVersion)
	assert.Equal(t, int64(0), cache.UpdatedAt)
}

func TestLocalStorage_ClearCache(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	cacheData := &CacheData{
		Entries:     []api.EntryResponse{},
		LastVersion: 1,
		UpdatedAt:   time.Now().Unix(),
	}

	err = storage.SaveCache(cacheData)
	require.NoError(t, err)

	err = storage.ClearCache()
	require.NoError(t, err)

	cache, err := storage.LoadCache()
	require.NoError(t, err)
	assert.Empty(t, cache.Entries)
	assert.Equal(t, int64(0), cache.LastVersion)
}

func TestLocalStorage_ClearCache_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: filepath.Join(tempDir, "token"),
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)
	err = storage.ClearCache()
	assert.NoError(t, err)
}

func TestLocalStorage_RelativeTokenPath(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.ClientConfig{
		DataDir:   tempDir,
		TokenFile: "token",
	}

	storage, err := NewLocalStorage(cfg)
	require.NoError(t, err)

	token := "test.jwt.token"

	err = storage.SaveToken(token)
	require.NoError(t, err)

	loadedToken, err := storage.LoadToken()
	require.NoError(t, err)
	assert.Equal(t, token, loadedToken)

	expectedPath := filepath.Join(tempDir, "token")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)
}
