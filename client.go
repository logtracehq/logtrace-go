package logtrace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.logtracehq.com/v1/developers"

// Client is the Logtrace API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	error      error
}

// New creates a Logtrace client. apiKey is required.
func New(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("logtrace: API key is required")
	}
	c := &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	return c, nil
}

func (c *Client) post(ctx context.Context, path string, body any) (*APIResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("logtrace: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultBaseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("logtrace: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("logtrace: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("logtrace: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &Error{StatusCode: resp.StatusCode}
		var res APIResponse
		if json.Unmarshal(respBody, &res) == nil {
			apiErr.Message = res.Message
		} else {
			apiErr.Message = string(respBody)
		}
		return nil, apiErr
	}

	var res APIResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("logtrace: failed to parse response: %w", err)
	}

	res.StatusCode = resp.StatusCode

	return &res, nil
}
