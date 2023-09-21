package envor

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// oncer for the config loading
var loadOnce sync.Once

// envOr returns the value of an environment variable if
// it exists, otherwise it returns the default value
func Get(name, def string) string {
	// Load .env file only once
	loadOnce.Do(func() {
		godotenv.Load(".env")
	})

	if value := os.Getenv(name); value != "" {
		return value
	}

	return def
}
