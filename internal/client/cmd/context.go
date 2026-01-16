package cmd

import (
	"github.com/Himany/GophKeeper/internal/client"
	"github.com/Himany/GophKeeper/internal/client/storage"
	"github.com/Himany/GophKeeper/internal/client/validator"
	"github.com/Himany/GophKeeper/internal/config"
)

type CommandContext struct {
	Config    *config.ClientConfig
	Storage   *storage.LocalStorage
	APIClient *client.Client
	Validator *validator.Validator
}

func NewCommandContext(cfg *config.ClientConfig) (*CommandContext, error) {
	clientStorage, err := storage.NewLocalStorage(cfg)
	if err != nil {
		return nil, err
	}

	apiClient := client.NewClient(cfg.ServerURL)

	if token, _ := clientStorage.LoadToken(); token != "" {
		apiClient.SetToken(token)
	}

	return &CommandContext{
		Config:    cfg,
		Storage:   clientStorage,
		APIClient: apiClient,
		Validator: validator.New(),
	}, nil
}
