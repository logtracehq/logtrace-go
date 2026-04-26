package logtrace

import (
	"context"
	"fmt"
	"time"
)

type Metadata map[string]any

type CreateEventRequest struct {
	ActionName      string   `json:"action_name"`
	UserID          string   `json:"user_id,omitempty"`
	Username        string   `json:"username,omitempty"`
	HTTPMethod      string   `json:"http_method"`
	HTTPStatus      string   `json:"http_status"`
	HTTPEndpoint    string   `json:"http_endpoint,omitempty"`
	ClientIP        string   `json:"client_ip"`
	ClientUserAgent string   `json:"client_user_agent"`
	Type            string   `json:"type,omitempty"`
	GeoIPLocation   string   `json:"geo_ip_location,omitempty"`
	Metadata        Metadata `json:"metadata,omitempty"`
}

type CreateSessionRequest struct {
	LoginAt    time.Time `json:"login_at"`
	Status     string    `json:"status"`
	UserID     string    `json:"user_id,omitempty"`
	Username   string    `json:"username,omitempty"`
	DeviceInfo string    `json:"device_info,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	Location   string    `json:"location,omitempty"`
	Token      string    `json:"token,omitempty"`
	Metadata   Metadata  `json:"metadata,omitempty"`
}

type CreateAuditLogRequest struct {
	Action    string   `json:"action"`
	Timestamp string   `json:"timestamp"`
	UserID    string   `json:"user_id,omitempty"`
	Username  string   `json:"username,omitempty"`
	IPAddress string   `json:"ip_address,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
	Metadata  Metadata `json:"metadata,omitempty"`
}

type APIResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("logtrace: %d - %s", e.StatusCode, e.Message)
}

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
