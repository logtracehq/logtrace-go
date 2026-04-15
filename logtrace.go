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
}

// Option configures the client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New creates a Logtrace client. apiKey is required.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// --- Request types ---

type CreateEventRequest struct {
	ActionName     string `json:"action_name"`
	UserID         string `json:"user_id,omitempty"`
	Username       string `json:"username,omitempty"`
	HTTPMethod     string `json:"http_method"`
	HTTPStatus     string `json:"http_status"`
	HTTPEndpoint   string `json:"http_endpoint,omitempty"`
	ClientIP       string `json:"client_ip"`
	ClientUserAgent string `json:"client_user_agent"`
	Type           string `json:"type,omitempty"`
	GeoIPLocation  string `json:"geo_ip_location,omitempty"`
}

type CreateSessionRequest struct {
	LoginAt    time.Time `json:"login_at"`
	Status     string    `json:"status"`
	UserID     string    `json:"user_id,omitempty"`
	Username   string    `json:"username,omitempty"`
	DeviceInfo string    `json:"device_info,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	Location   string    `json:"location,omitempty"`
}

type Metadata struct {
	Event       string `json:"event,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreateAuditLogRequest struct {
	Action    string    `json:"action"`
	Timestamp string    `json:"timestamp"`
	UserID    string    `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Metadata  *Metadata `json:"metadata,omitempty"`
}

// --- Response types ---

type APIResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

// Error represents an API error.
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("logtrace: %d - %s", e.StatusCode, e.Message)
}

// --- API methods ---

// CreateEvent sends an event to Logtrace.
func (c *Client) CreateEvent(ctx context.Context, req *CreateEventRequest) (*APIResponse, error) {
	return c.post(ctx, "/events", req)
}

// CreateSession sends a session to Logtrace.
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*APIResponse, error) {
	return c.post(ctx, "/sessions", req)
}

// CreateAuditLog sends an audit log to Logtrace.
func (c *Client) CreateAuditLog(ctx context.Context, req *CreateAuditLogRequest) (*APIResponse, error) {
	return c.post(ctx, "/audit-logs", req)
}

// --- internal ---

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
		var apiResp APIResponse
		if json.Unmarshal(respBody, &apiResp) == nil {
			apiErr.Message = apiResp.Message
		} else {
			apiErr.Message = string(respBody)
		}
		return nil, apiErr
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("logtrace: failed to parse response: %w", err)
	}
	apiResp.StatusCode = resp.StatusCode

	return &apiResp, nil
}
