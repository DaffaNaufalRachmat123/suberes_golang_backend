// Package sentryutil wraps the Sentry SDK initialisation.
// If SENTRY_DSN is not set the package is a no-op, so development keeps working.
package sentryutil

import (
	"os"
	"time"

	"suberes_golang/logger"

	"github.com/getsentry/sentry-go"
)

// Init initialises the Sentry SDK.
// No-op (with a warning log) when SENTRY_DSN is not set.
func Init() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		logger.Logger.Warn().Msg("SENTRY_DSN not configured — Sentry error tracking disabled")
		return
	}

	env := os.Getenv("APP_ENV")
	release := os.Getenv("APP_VERSION")
	if release == "" {
		release = "unknown"
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		Release:          release,
		AttachStacktrace: true,
		// Sample 10 % of transactions for performance monitoring.
		TracesSampleRate: 0.1,
	})
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Sentry initialisation failed")
		return
	}

	logger.Logger.Info().
		Str("env", env).
		Str("release", release).
		Msg("Sentry initialised")
}

// Flush blocks until all buffered events are sent to Sentry, or 2 s elapses.
// Call this in the graceful-shutdown block.
func Flush() {
	sentry.Flush(2 * time.Second)
}

// CaptureError sends a non-fatal error to Sentry with additional context.
func CaptureError(err error) {
	sentry.CaptureException(err)
}
