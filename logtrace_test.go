package logtrace

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testClient(fn roundTripFunc) (*Client, error) {
	client, err := New("test-api-key")
	if err != nil {
		slog.Error("failed to create client", "error", err)
	}
	client.httpClient.Transport = fn
	return client, nil
}

func testRequestClient(c *Client) *requestClient {
	return &requestClient{
		client:    c,
		method:    "POST",
		endpoint:  "/run",
		clientIP:  "192.168.1.1",
		userAgent: "TestAgent/1.0",
		status:    func() *int { s := 200; return &s }(),
	}
}

func jsonResponse(status int, body map[string]any) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(b))),
	}
}

func TestCreateEvent_Success(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, map[string]any{
			"message":    "Event created",
			"statusCode": 200,
		}), nil
	})

	lc := testRequestClient(client)
	resp, err := lc.CreateEvent(context.Background(), &CreateEventRequest{
		ActionName: "user.login",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Message != "Event created" {
		t.Errorf("expected message 'Event created', got %q", resp.Message)
	}
}

func TestCreateEvent_MiddlewareFieldsInjected(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]any
		json.Unmarshal(body, &data)

		if data["action_name"] != "user.signup" {
			t.Errorf("expected action_name 'user.signup', got %v", data["action_name"])
		}
		if data["http_method"] != "POST" {
			t.Errorf("expected injected http_method 'POST', got %v", data["http_method"])
		}
		if data["client_ip"] != "192.168.1.1" {
			t.Errorf("expected injected client_ip '192.168.1.1', got %v", data["client_ip"])
		}
		if data["client_user_agent"] != "TestAgent/1.0" {
			t.Errorf("expected injected user_agent 'TestAgent/1.0', got %v", data["client_user_agent"])
		}
		if data["http_endpoint"] != "/run" {
			t.Errorf("expected injected endpoint '/run', got %v", data["http_endpoint"])
		}
		if data["http_status"] != float64(200) {
			t.Errorf("expected injected http_status 200, got %v", data["http_status"])
		}
		if _, ok := data["user_id"]; ok {
			t.Error("expected user_id to be omitted when empty")
		}

		return jsonResponse(200, map[string]any{"message": "ok", "statusCode": 200}), nil
	})

	lc := testRequestClient(client)
	lc.CreateEvent(context.Background(), &CreateEventRequest{
		ActionName: "user.signup",
	})
}

func TestCreateEvent_WithOptionalFields(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]any
		json.Unmarshal(body, &data)

		if data["user_id"] != "usr_123" {
			t.Errorf("expected user_id 'usr_123', got %v", data["user_id"])
		}
		if data["username"] != "john" {
			t.Errorf("expected username 'john', got %v", data["username"])
		}
		if data["geo_ip_location"] != "US" {
			t.Errorf("expected geo_ip_location 'US', got %v", data["geo_ip_location"])
		}

		return jsonResponse(200, map[string]any{"message": "ok", "statusCode": 200}), nil
	})

	lc := testRequestClient(client)
	lc.CreateEvent(context.Background(), &CreateEventRequest{
		ActionName:    "user.login",
		UserID:        "usr_123",
		UserName:      "john",
		GeoIPLocation: "US",
	})
}

func TestCreateSession_Success(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, map[string]any{
			"message":    "Session created",
			"statusCode": 200,
		}), nil
	})

	lc := testRequestClient(client)
	resp, err := lc.CreateSession(context.Background(), &CreateSessionRequest{
		LoginAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Status:  "ACTIVE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Session created" {
		t.Errorf("expected message 'Session created', got %q", resp.Message)
	}
}

func TestCreateSession_WithAllFields(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]any
		json.Unmarshal(body, &data)

		if data["user_id"] != "usr_456" {
			t.Errorf("expected user_id 'usr_456', got %v", data["user_id"])
		}
		if data["device_info"] != "Chrome on macOS" {
			t.Errorf("expected device_info 'Chrome on macOS', got %v", data["device_info"])
		}
		if data["ip_address"] != "10.0.0.5" {
			t.Errorf("expected ip_address '10.0.0.5', got %v", data["ip_address"])
		}

		return jsonResponse(200, map[string]any{"message": "ok", "statusCode": 200}), nil
	})

	lc := testRequestClient(client)
	lc.CreateSession(context.Background(), &CreateSessionRequest{
		LoginAt:    time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC),
		LogoutAt:   time.Now().AddDate(0, 0, 1),
		Status:     "ACTIVE",
		UserID:     "usr_456",
		UserName:   "jane",
		DeviceInfo: "Chrome on macOS",
		IPAddress:  "10.0.0.5",
		Location:   "New York, US",
	})
}

func TestCreateAuditLog_Success(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, map[string]any{
			"message":    "Audit log created",
			"statusCode": 200,
		}), nil
	})

	lc := testRequestClient(client)
	resp, err := lc.CreateAuditLog(context.Background(), &CreateAuditLogRequest{
		Action:    "user.deleted",
		Timestamp: "2025-03-10T14:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Audit log created" {
		t.Errorf("expected message 'Audit log created', got %q", resp.Message)
	}
}

func TestCreateAuditLog_WithMetadata(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]any
		json.Unmarshal(body, &data)

		meta, ok := data["metadata"].(map[string]any)
		if !ok {
			t.Fatal("expected metadata to be present")
		}
		if meta["event"] != "role_change" {
			t.Errorf("expected metadata.event 'role_change', got %v", meta["event"])
		}
		if meta["description"] != "Promoted to admin" {
			t.Errorf("expected metadata.description 'Promoted to admin', got %v", meta["description"])
		}

		return jsonResponse(200, map[string]any{"message": "ok", "statusCode": 200}), nil
	})

	lc := testRequestClient(client)
	lc.CreateAuditLog(context.Background(), &CreateAuditLogRequest{
		Action:    "user.role_change",
		Timestamp: "2025-03-10T14:00:00Z",
		UserID:    "usr_789",
		RequestID: "req_abc",
		Metadata: Metadata{
			"event":       "role_change",
			"description": "Promoted to admin",
		},
	})
}

func TestPost_SendsCorrectHeaders(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", got)
		}
		if got := r.Header.Get("X-API-Key"); got != "test-api-key" {
			t.Errorf("expected X-API-Key 'test-api-key', got %q", got)
		}
		return jsonResponse(200, map[string]any{"message": "ok", "statusCode": 200}), nil
	})

	testRequestClient(client).CreateEvent(context.Background(), &CreateEventRequest{
		ActionName: "test",
	})
}

func TestPost_SendsToCorrectEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*requestClient) error
		wantPath string
	}{
		{
			name: "events",
			call: func(lc *requestClient) error {
				_, err := lc.CreateEvent(context.Background(), &CreateEventRequest{
					ActionName: "t",
				})
				return err
			},
			wantPath: "/v1/developers/events",
		},
		{
			name: "sessions",
			call: func(lc *requestClient) error {
				_, err := lc.CreateSession(context.Background(), &CreateSessionRequest{
					LoginAt:  time.Now(),
					LogoutAt: time.Now().AddDate(0, 0, 7),
					Status:   "active",
				})
				return err
			},
			wantPath: "/v1/developers/sessions",
		},
		{
			name: "audit-logs",
			call: func(lc *requestClient) error {
				_, err := lc.CreateAuditLog(context.Background(), &CreateAuditLogRequest{
					Action: "t", Timestamp: "2025-01-01T00:00:00Z",
				})
				return err
			},
			wantPath: "/v1/developers/audit-logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := testClient(func(r *http.Request) (*http.Response, error) {
				if !strings.HasSuffix(r.URL.Path, tt.wantPath) {
					t.Errorf("expected path ending in %q, got %q", tt.wantPath, r.URL.Path)
				}
				return jsonResponse(200, map[string]any{"message": "ok", "statusCode": 200}), nil
			})
			tt.call(testRequestClient(client))
		})
	}
}

// ---- Error handling tests ----

func TestPost_APIError_400(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(400, map[string]any{
			"message":    "Bad request: missing action_name",
			"statusCode": 400,
		}), nil
	})

	lc := testRequestClient(client)
	_, err := lc.CreateEvent(context.Background(), &CreateEventRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Message, "missing action_name") {
		t.Errorf("expected message to contain 'missing action_name', got %q", apiErr.Message)
	}
}

func TestPost_APIError_401_Unauthorized(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(401, map[string]any{
			"message":    "Invalid API key",
			"statusCode": 401,
		}), nil
	})

	lc := testRequestClient(client)
	_, err := lc.CreateEvent(context.Background(), &CreateEventRequest{ActionName: "t"})

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestPost_APIError_500_ServerError(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(500, map[string]any{
			"message":    "Internal server error",
			"statusCode": 500,
		}), nil
	})

	lc := testRequestClient(client)
	_, err := lc.CreateAuditLog(context.Background(), &CreateAuditLogRequest{
		Action: "t", Timestamp: "2025-01-01T00:00:00Z",
	})

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestPost_APIError_NonJSONResponse(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 502,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(strings.NewReader("Bad Gateway")),
		}, nil
	})

	lc := testRequestClient(client)
	_, err := lc.CreateEvent(context.Background(), &CreateEventRequest{ActionName: "t"})

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != 502 {
		t.Errorf("expected status 502, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Bad Gateway" {
		t.Errorf("expected message 'Bad Gateway', got %q", apiErr.Message)
	}
}

func TestPost_NetworkError(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return nil, &net_error{msg: "connection refused"}
	})

	lc := testRequestClient(client)
	_, err := lc.CreateEvent(context.Background(), &CreateEventRequest{ActionName: "t"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("expected error to contain 'request failed', got %q", err.Error())
	}
}

func TestFromContext_ReturnsInjectedClient(t *testing.T) {
	base, _ := New("key")
	rc := testRequestClient(base)
	ctx := context.WithValue(context.Background(), clientKey, rc)

	got := FromContext(ctx, base)
	if got != rc {
		t.Error("expected FromContext to return the injected requestClient")
	}
}

func TestFromContext_FallsBackToBaseClient(t *testing.T) {
	base, _ := New("key")
	got := FromContext(context.Background(), base)
	if got.client != base {
		t.Error("expected FromContext fallback to wrap the base client")
	}
}

func TestNew_DefaultHTTPClient(t *testing.T) {
	c, _ := New("my-key")
	if c.apiKey != "my-key" {
		t.Errorf("expected apiKey 'my-key', got %q", c.apiKey)
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
	if c.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", c.httpClient.Timeout)
	}
}

func TestError_ErrorString(t *testing.T) {
	e := &Error{StatusCode: 404, Message: "Not found"}
	want := "logtrace: 404 - Not found"
	if e.Error() != want {
		t.Errorf("expected %q, got %q", want, e.Error())
	}
}

func TestPost_CancelledContext(t *testing.T) {
	client, _ := testClient(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	lc := testRequestClient(client)
	_, err := lc.CreateEvent(ctx, &CreateEventRequest{ActionName: "t"})

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

type net_error struct{ msg string }

func (e *net_error) Error() string   { return e.msg }
func (e *net_error) Timeout() bool   { return false }
func (e *net_error) Temporary() bool { return false }
