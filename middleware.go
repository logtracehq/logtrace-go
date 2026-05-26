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

func (c *Client) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := newResponseWriter(w)
		status := http.StatusOK

		rc := &requestClient{
			client:    c,
			method:    r.Method,
			endpoint:  r.URL.RequestURI(),
			clientIP:  realIP(r),
			userAgent: r.UserAgent(),
			status:    &status,
		}

		ctx := context.WithValue(r.Context(), clientKey, rc)
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		*rc.status = wrapped.statusCode
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
