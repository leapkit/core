package envor

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

// oncer for the config loading
var loadOnce sync.Once

func init() {
	loadOnce.Do(loadENV)
}

// envOr returns the value of an environment variable if
// it exists, otherwise it returns the default value
func Get(name, def string) string {

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

	for key, value := range parseVars(file) {
		err := os.Setenv(key, value)
		if err != nil {
			continue
		}
	}
}

// parseVars reads the variables from the reader and sets them
// in the environment.
func parseVars(r io.Reader) map[string]string {
	vars := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		value = strings.Trim(value, "\"")

		vars[key] = value
	}

	return vars
}
