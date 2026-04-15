# logtrace-go

Go SDK for the [Logtrace](https://logtracehq.com) developer API.

## Install

```bash
go get github.com/logtracehq/logtrace-go
```

## Usage

```go
package main

import (
	"context"
	"log"
	"time"

	logtrace "github.com/logtracehq/logtrace-go"
)

func main() {
	client := logtrace.New("your-api-key")

	// Create an event
	_, err := client.CreateEvent(context.Background(), &logtrace.CreateEventRequest{
		ActionName:      "user.login",
		Username:        "jane_doe",
		HTTPMethod:      "POST",
		HTTPStatus:      "200",
		ClientIP:        "192.168.1.1",
		ClientUserAgent: "Mozilla/5.0",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a session
	_, err = client.CreateSession(context.Background(), &logtrace.CreateSessionRequest{
		LoginAt:  time.Now(),
		Status:   "ACTIVE",
		Username: "jane_doe",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create an audit log
	_, err = client.CreateAuditLog(context.Background(), &logtrace.CreateAuditLogRequest{
		Action:    "user.deleted",
		Timestamp: time.Now().Format(time.RFC3339),
		Username:  "jane_doe",
		Metadata: &logtrace.Metadata{
			Event:       "deletion",
			Type:        "user",
			Description: "User account was deleted",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Custom Configuration

```go
client := logtrace.New("your-api-key",
	logtrace.WithBaseURL("https://your-instance.com/v1/developers"),
	logtrace.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```
