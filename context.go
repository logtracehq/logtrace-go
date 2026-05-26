package logtrace

import "context"

type contextKey string

const clientKey contextKey = "logtrace_client"

type requestClient struct {
	client    *Client
	method    string
	endpoint  string
	clientIP  string
	userAgent string
	status    *int
}

func (rc *requestClient) CreateEvent(ctx context.Context, req *CreateEventRequest) (*APIResponse, error) {
	req.HTTPMethod = rc.method
	req.HTTPEndpoint = rc.endpoint
	req.ClientIP = rc.clientIP
	req.ClientUserAgent = rc.userAgent
	if rc.status != nil {
		req.HTTPStatus = *rc.status
	}
	return rc.client.CreateEvent(ctx, req)
}

func (rc *requestClient) CreateSession(ctx context.Context, req *CreateSessionRequest) (*APIResponse, error) {
	return rc.client.CreateSession(ctx, req)
}

func (rc *requestClient) CreateAuditLog(ctx context.Context, req *CreateAuditLogRequest) (*APIResponse, error) {
	return rc.client.CreateAuditLog(ctx, req)
}

func FromContext(ctx context.Context, fallback *Client) *requestClient {
	if rc, ok := ctx.Value(clientKey).(*requestClient); ok {
		return rc
	}

	return &requestClient{client: fallback}
}
