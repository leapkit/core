// envload package loads .env files into the environment. To do it
// it uses an init function that reads the .env file and sets the
// variables in the environment.
package envload

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// init reads the .env file and sets the variables in the environment
// and passes it to the parseVars function, then it set the variables
// in the environment.
func init() {
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
