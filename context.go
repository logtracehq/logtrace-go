package logtrace

import (
	"context"
	"time"
)

type contextKey string

const clientKey contextKey = "logtrace_client"

type requestClient struct {
	client          *Client
	method          string
	headers         map[string]string
	endpoint        string
	clientIP        string
	userAgent       string
	status          *int
	operatingSystem string
}

func (rc *requestClient) buildRequestDetails() RequestDetails {
	var status int
	if rc.status != nil {
		status = *rc.status
	}

	return RequestDetails{
		Timestamp:       time.Now().UTC(),
		HTTPMethod:      rc.method,
		HTTPEndpoint:    rc.endpoint,
		IPAddress:       rc.clientIP,
		ClientUserAgent: rc.userAgent,
		HTTPStatusCode:  status,
		OperatingSystem: rc.operatingSystem,
		RequestHeaders:  rc.headers,
	}
}

func (rc *requestClient) CreateEvent(ctx context.Context, req *CreateEventRequest) (*APIResponse, error) {
	req.RequestDetails = rc.buildRequestDetails()
	return rc.client.CreateEvent(ctx, req)
}

func (rc *requestClient) CreateSession(ctx context.Context, req *CreateSessionRequest) (*APIResponse, error) {
	req.RequestDetails = rc.buildRequestDetails()

	return rc.client.CreateSession(ctx, req)
}

func (rc *requestClient) CreateAuditLog(ctx context.Context, req *CreateAuditLogRequest) (*APIResponse, error) {
	req.RequestDetails = rc.buildRequestDetails()

	return rc.client.CreateAuditLog(ctx, req)
}

func FromContext(ctx context.Context, fallback *Client) *requestClient {
	if rc, ok := ctx.Value(clientKey).(*requestClient); ok {
		return rc
	}

	return &requestClient{client: fallback}
}
