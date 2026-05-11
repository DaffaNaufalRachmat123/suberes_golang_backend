package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is the global structured logger instance.
var Logger zerolog.Logger

// Init initialises the global zerolog logger.
// Call this once, as early as possible in main(), after loading .env.
func Init() {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))

	var writer io.Writer
	if env == "dev" || env == "" {
		// Human-readable console output for local development.
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	} else {
		// JSON output for staging/production — easy to ingest by log aggregators.
		writer = os.Stdout
	}

	level := zerolog.InfoLevel
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		level = zerolog.DebugLevel
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "unknown"
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(level)

	Logger = zerolog.New(writer).With().
		Timestamp().
		Str("service", "suberes-api").
		Str("env", env).
		Str("version", version).
		Logger()

	// Align the global zerolog default logger as well (used by third-party libs).
	log.Logger = Logger
}

// Get returns a pointer to the global logger (convenience accessor).
func Get() *zerolog.Logger {
	return &Logger
}
