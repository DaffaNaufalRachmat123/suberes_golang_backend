// Package appenv validates required environment variables at application startup.
// If any required variable is missing the application panics with a clear,
// actionable message before any side-effectful initialisation takes place.
package appenv

import (
	"fmt"
	"os"
	"strings"
)

// required lists the environment variables that MUST be set before the app starts.
// DB variables use a prefix (DEV_ / STAG_ / PROD_) that is determined by APP_ENV,
// so only the universal variables are listed here; DB connectivity is validated
// indirectly by config.ConnectDB().
var required = []string{
	"APP_ENV",
	"SECRET_KEY",
	"SECRET_KEY_REFRESH",
	"REDIS_HOST",
	"REDIS_PORT",
}

// Validate checks every variable in `required` and panics if any are missing.
// Call this at the very top of main(), immediately after godotenv.Load().
func Validate() {
	missing := make([]string, 0, len(required))
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		panic(fmt.Sprintf(
			"\n[STARTUP FAILURE] Missing required environment variables:\n  %s\n\n"+
				"Configure them in your .env file or environment before starting the application.\n"+
				"See .env.example for reference.\n",
			strings.Join(missing, "\n  "),
		))
	}
}
