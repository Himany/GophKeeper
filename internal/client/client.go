package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Himany/GophKeeper/pkg/api"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type Client struct {
	httpClient *resty.Client
	baseURL    string
	token      string
}

func NewClient(baseURL string) *Client {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return &Client{
		httpClient: client,
		baseURL:    baseURL,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
	c.httpClient.SetAuthToken(token)
}

func (c *Client) GetToken() string {
	return c.token
}

func (c *Client) Register(ctx context.Context, username, password string) (*api.AuthResponse, error) {
	req := api.RegisterRequest{
		Username: username,
		Password: password,
	}

	var response api.AuthResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Post(c.baseURL + "/api/v1/register")

	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("registration failed: %s", errResp.Message)
	}

	c.SetToken(response.Token)

	return &response, nil
}

func (c *Client) Login(ctx context.Context, username, password string) (*api.AuthResponse, error) {
	req := api.LoginRequest{
		Username: username,
		Password: password,
	}

	var response api.AuthResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Post(c.baseURL + "/api/v1/login")

	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("login failed: %s", errResp.Message)
	}

	c.SetToken(response.Token)

	return &response, nil
}

func (c *Client) CreateEntry(ctx context.Context, name, entryType string, data map[string]string, metadata string) (*api.EntryResponse, error) {
	req := api.CreateEntryRequest{
		Name:     name,
		Type:     entryType,
		Data:     data,
		Metadata: metadata,
	}

	var response api.EntryResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Post(c.baseURL + "/api/v1/entries")

	if err != nil {
		return nil, fmt.Errorf("create entry request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("create entry failed: %s", errResp.Message)
	}

	return &response, nil
}

func (c *Client) GetEntry(ctx context.Context, entryID uuid.UUID) (*api.EntryResponse, error) {
	var response api.EntryResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Get(c.baseURL + "/api/v1/entries/" + entryID.String())

	if err != nil {
		return nil, fmt.Errorf("get entry request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("get entry failed: %s", errResp.Message)
	}

	return &response, nil
}

func (c *Client) ListEntries(ctx context.Context) (*api.ListEntriesResponse, error) {
	var response api.ListEntriesResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Get(c.baseURL + "/api/v1/entries")

	if err != nil {
		return nil, fmt.Errorf("list entries request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("list entries failed: %s", errResp.Message)
	}

	return &response, nil
}

func (c *Client) UpdateEntry(ctx context.Context, entryID uuid.UUID, name string, data map[string]string, metadata string) (*api.EntryResponse, error) {
	req := api.UpdateEntryRequest{
		Name:     name,
		Data:     data,
		Metadata: metadata,
	}

	var response api.EntryResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Put(c.baseURL + "/api/v1/entries/" + entryID.String())

	if err != nil {
		return nil, fmt.Errorf("update entry request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("update entry failed: %s", errResp.Message)
	}

	return &response, nil
}

func (c *Client) DeleteEntry(ctx context.Context, entryID uuid.UUID) error {
	var response api.SuccessResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Delete(c.baseURL + "/api/v1/entries/" + entryID.String())

	if err != nil {
		return fmt.Errorf("delete entry request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return fmt.Errorf("delete entry failed: %s", errResp.Message)
	}

	return nil
}

func (c *Client) Sync(ctx context.Context, lastVersion int64) (*api.SyncResponse, error) {
	req := api.SyncRequest{
		LastVersion: lastVersion,
	}

	var response api.SyncResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetError(&api.ErrorResponse{}).
		Post(c.baseURL + "/api/v1/sync")

	if err != nil {
		return nil, fmt.Errorf("sync request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return nil, fmt.Errorf("sync failed: %s", errResp.Message)
	}

	return &response, nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get(c.baseURL + "/api/v1/health")

	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("server is not healthy: status %d", resp.StatusCode())
	}

	return nil
}
