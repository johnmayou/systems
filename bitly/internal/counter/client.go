package counter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	url  string
	http http.Client
}

func NewClient(url string) *Client {
	return &Client{url: url, http: http.Client{}}
}

func (c *Client) GetCount(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return 0, fmt.Errorf("counter: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("counter: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("counter: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Value int `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("counter: decode response: %w", err)
	}

	return body.Value, nil
}
