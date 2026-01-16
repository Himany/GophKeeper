package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Himany/GophKeeper/internal/config"
	"github.com/Himany/GophKeeper/pkg/api"
)

type LocalStorage struct {
	dataDir   string
	tokenFile string
}

func NewLocalStorage(cfg *config.ClientConfig) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &LocalStorage{
		dataDir:   cfg.DataDir,
		tokenFile: cfg.TokenFile,
	}, nil
}

func (s *LocalStorage) SaveToken(token string) error {
	tokenPath := s.tokenFile
	if !filepath.IsAbs(tokenPath) {
		tokenPath = filepath.Join(s.dataDir, tokenPath)
	}

	if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

func (s *LocalStorage) LoadToken() (string, error) {
	tokenPath := s.tokenFile
	if !filepath.IsAbs(tokenPath) {
		tokenPath = filepath.Join(s.dataDir, tokenPath)
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to load token: %w", err)
	}

	return string(data), nil
}

func (s *LocalStorage) RemoveToken() error {
	tokenPath := s.tokenFile
	if !filepath.IsAbs(tokenPath) {
		tokenPath = filepath.Join(s.dataDir, tokenPath)
	}

	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token: %w", err)
	}

	return nil
}

type CacheData struct {
	Entries     []api.EntryResponse `json:"entries"`
	LastVersion int64               `json:"last_version"`
	UpdatedAt   int64               `json:"updated_at"`
}

func (s *LocalStorage) SaveCache(data *CacheData) error {
	cachePath := filepath.Join(s.dataDir, "cache.json")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(cachePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}

func (s *LocalStorage) LoadCache() (*CacheData, error) {
	cachePath := filepath.Join(s.dataDir, "cache.json")

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &CacheData{}, nil
		}
		return nil, fmt.Errorf("failed to load cache: %w", err)
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &cacheData, nil
}

func (s *LocalStorage) ClearCache() error {
	cachePath := filepath.Join(s.dataDir, "cache.json")

	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}
