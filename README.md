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
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"

	logtrace "github.com/logtracehq/logtrace-go"
)

func main() {
	client := logtrace.New(os.Getenv("API_KEY"))

	mux := http.NewServeMux()
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		lc := logtrace.FromContext(r.Context(), client)

		_, err := lc.CreateEvent(r.Context(), &logtrace.CreateEventRequest{
			Name: "user.login",
			UserID:     "12345",
			UserName:   "jane_doe",
			HTTPStatus: "200",
			Metadata: logtrace.M    etadata{
				"name":      "login",
				"type":        "user",
				"description": "User logged in successfully",
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a session
		_, err = lc.CreateSession(r.Context(), &logtrace.CreateSessionRequest{
			LoginAt:  time.Now(),
			Status:   "active",
			UserName: "jane_doe",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create an audit log
		_, err = lc.CreateAuditLog(r.Context(), &logtrace.CreateAuditLogRequest{
			Name:    "user.deleted",
			Timestamp: time.Now().Format(time.RFC3339),
			UserName:  "jane_doe",
			Metadata: logtrace.Metadata{
				"name":      "deletion",
				"type":        "user",
				"description": "User account was deleted",
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("done"))
	})

	log.Println("Server running on :5000")
	log.Fatal(http.ListenAndServe(":5000", client.Logger(mux)))
}
```
