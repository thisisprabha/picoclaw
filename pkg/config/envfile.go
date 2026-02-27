package config

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// LoadEnvFile reads a dotenv-like file and sets variables in the process env.
// Supported formats:
// - KEY=value
// - export KEY=value
// - blank lines and comments (#...)
//
// When overwrite=false, existing process env vars are preserved.
// Returns loaded=true when the file exists and was processed.
func LoadEnvFile(path string, overwrite bool) (loaded bool, err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return false, nil
	}
	path = expandHome(path)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		if !overwrite {
			if _, exists := os.LookupEnv(key); exists {
				continue
			}
		}
		if setErr := os.Setenv(key, value); setErr != nil {
			return true, setErr
		}
	}
	if err := scanner.Err(); err != nil {
		return true, err
	}
	return true, nil
}

func parseEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	}

	eq := strings.Index(line, "=")
	if eq <= 0 {
		return "", "", false
	}

	key = strings.TrimSpace(line[:eq])
	if !envKeyPattern.MatchString(key) {
		return "", "", false
	}

	value = strings.TrimSpace(line[eq+1:])
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if first == '"' && last == '"' {
			unquoted, err := strconv.Unquote(value)
			if err == nil {
				value = unquoted
			} else {
				value = value[1 : len(value)-1]
			}
		} else if first == '\'' && last == '\'' {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true
}
