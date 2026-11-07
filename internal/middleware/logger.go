package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type userHolderKeyType struct{}

var userHolderKey = userHolderKeyType{}

type userHolder struct {
	email string
	role  string
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// RequestLogger is an enriched contextual logging middleware.
// In development, it prints concise, human-friendly colored status lines with authenticated user context.
// In production, it emits structured JSON logs.
func RequestLogger(env string) func(next http.Handler) http.Handler {
	isDev := env == "development" || env == "dev" || env == ""

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := chiMiddleware.GetReqID(r.Context())
			if reqID == "" {
				reqID = "-"
			}

			holder := &userHolder{}
			ctxWithHolder := context.WithValue(r.Context(), userHolderKey, holder)
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r.WithContext(ctxWithHolder))

			duration := time.Since(start)

			userContextStr := ""
			if holder.email != "" {
				userContextStr = fmt.Sprintf("[%s | %s]", holder.email, holder.role)
			}

			if isDev {
				statusEmoji := "🟢"
				switch {
				case rec.status >= 500:
					statusEmoji = "🔥"
				case rec.status >= 400:
					statusEmoji = "🔴"
				case rec.status >= 300:
					statusEmoji = "🟡"
				}

				if userContextStr != "" {
					fmt.Printf("%s %d %-6s %s %s (%s, %dB)\n", statusEmoji, rec.status, r.Method, r.URL.RequestURI(), userContextStr, duration.Round(time.Microsecond), rec.bytes)
				} else {
					fmt.Printf("%s %d %-6s %s (%s, %dB)\n", statusEmoji, rec.status, r.Method, r.URL.RequestURI(), duration.Round(time.Microsecond), rec.bytes)
				}
			} else {
				level := slog.LevelInfo
				if rec.status >= 500 {
					level = slog.LevelError
				} else if rec.status >= 400 {
					level = slog.LevelWarn
				}

				attrs := []any{
					"request_id", reqID,
					"method", r.Method,
					"path", r.URL.Path,
					"status", rec.status,
					"duration_ms", duration.Milliseconds(),
					"bytes", rec.bytes,
					"remote_ip", r.RemoteAddr,
				}
				if holder.email != "" {
					attrs = append(attrs, "email", holder.email, "role", holder.role)
				}

				slog.Log(r.Context(), level, "HTTP request handled", attrs...)
			}
		})
	}
}
