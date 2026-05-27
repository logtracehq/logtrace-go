package logtrace

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type RequestLog struct {
	Timestamp  string            `json:"timestamp"`
	Method     string            `json:"method"`
	Endpoint   string            `json:"endpoint"`
	IPAddress  string            `json:"ip_address"`
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"status_code"`
	Duration   string            `json:"duration"`
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logs your requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		headers := make(map[string]string, len(r.Header))
		for key, values := range r.Header {
			headers[key] = values[len(values)-1]
		}

		entry := RequestLog{
			Timestamp:  start.UTC().Format(time.RFC3339),
			Method:     r.Method,
			Endpoint:   r.URL.RequestURI(),
			IPAddress:  realIP(r),
			Headers:    headers,
			StatusCode: wrapped.statusCode,
			Duration:   time.Since(start).String(),
		}

		logJSON(entry)
	})
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if i := len(ip); i > 0 {
			for j := 0; j < i; j++ {
				if ip[j] == ',' {
					return strings.TrimSpace(ip[:j])
				}
			}
		}
		return strings.TrimSpace(ip)
	}
	return r.RemoteAddr
}

func logJSON(entry RequestLog) {
	b, err := json.Marshal(entry)
	if err != nil {
		log.Printf("middleware: failed to marshal log entry: %v", err)
		return
	}
	log.Println(string(b))
}
