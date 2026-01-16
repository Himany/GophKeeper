package client

import (
	"context"
	"fmt"

	"github.com/Himany/GophKeeper/pkg/api"
	"github.com/go-resty/resty/v2"
)

func doRequest[T any](
	ctx context.Context,
	client *resty.Client,
	method, url string,
	body interface{},
	result *T,
) error {
	req := client.R().
		SetContext(ctx).
		SetResult(result).
		SetError(&api.ErrorResponse{})

	if body != nil {
		req.SetBody(body)
	}

	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		errResp := resp.Error().(*api.ErrorResponse)
		return fmt.Errorf("API error: %s", errResp.Message)
	}

	return nil
}
