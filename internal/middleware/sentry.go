package middleware

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// SentryMiddleware returns an http.Handler middleware that reports panics and errors to Sentry.
// If Sentry is not initialized (DSN empty), it's a no-op passthrough.
func SentryMiddleware() func(http.Handler) http.Handler {
	if sentry.CurrentHub().Client() == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	handler := sentryhttp.New(sentryhttp.Options{
		Repanic: true, // let the Recoverer middleware handle the response
	})
	return handler.Handle
}
