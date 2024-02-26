package envor

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// oncer for the config loading
var loadOnce sync.Once

// envOr returns the value of an environment variable if
// it exists, otherwise it returns the default value
func Get(name, def string) string {
	loadOnce.Do(loadENV)
	if value := os.Getenv(name); value != "" {
		return value
	}

	return def
}

// loadEnv loads the .env file into the environment
// variables it only loads that file.
func loadENV() {
	// open .env file
	file, err := os.Open(".env")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		pair := strings.Split(line, "=")
		if len(pair) != 2 {
			continue
		}

		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		value = strings.Trim(value, "\"")

		err := os.Setenv(key, value)
		if err != nil {
			continue
		}
	}
}
