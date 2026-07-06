package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AdventurerAmer/shortner/errs"
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
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("'client.Do' failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var expectedErr errs.Error
		_ = json.Unmarshal(body, &expectedErr)
		return fmt.Errorf("request failed with status %q, (%d): %w", resp.Status, resp.StatusCode, expectedErr)
	}
	return nil
}
