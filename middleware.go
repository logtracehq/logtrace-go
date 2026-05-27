package logtrace

import (
	"context"
	"net/http"
	"strings"
)

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

// Handles injection of request details
func Logger(client *Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped := newResponseWriter(w)

			userAgent := r.UserAgent()

			rc := &requestClient{
				client:          client,
				method:          r.Method,
				endpoint:        r.URL.RequestURI(),
				clientIP:        realIP(r),
				userAgent:       userAgent,
				status:          &wrapped.statusCode,
				operatingSystem: operatingSystem(userAgent),
			}

			ctx := context.WithValue(r.Context(), clientKey, rc)
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			headers := make(map[string]string, len(r.Header))
			for key, values := range r.Header {
				headers[key] = values[len(values)-1]
			}

			rc.headers = headers
		})
	}
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

func operatingSystem(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "curl"):
		return "Unknown (curl)"

	case strings.Contains(ua, "windows"):
		return "Windows"

	case strings.Contains(ua, "mac os"),
		strings.Contains(ua, "macintosh"),
		strings.Contains(ua, "darwin"):
		return "macOS"

	case strings.Contains(ua, "android"):
		return "Android"

	case strings.Contains(ua, "iphone"),
		strings.Contains(ua, "ipad"),
		strings.Contains(ua, "ios"):
		return "iOS"

	case strings.Contains(ua, "linux"):
		return "Linux"

	case strings.Contains(ua, "cros"):
		return "Chrome OS"

	default:
		return "Unknown"
	}
}
