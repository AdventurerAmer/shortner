package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Address string
}

func NewClient(address string) *Client {
	return &Client{Address: address}
}

func (c *Client) IncrementClicks(ctx context.Context, alias string) error {
	endpoint := fmt.Sprintf("%s/v1/analytics/%s/clicks", c.Address, alias)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("'http.NewRequestWithContext' failed: %w", err)
	}
	r.Header.Set("Accept", "application/json")
	client := &http.Client{
		Timeout: time.Second,
	}
	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("'client.Do' failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %q, code: %d", resp.Status, resp.StatusCode)
	}
	return nil
}
