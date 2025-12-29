package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetEnv returns an environment variable value or a default.
func GetEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvRequired returns an environment variable or an error if not set.
func GetEnvRequired(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", ErrConfigRequired("GetEnvRequired", key)
	}
	return value, nil
}

// GetEnvInt returns an environment variable as an int or a default.
func GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvIntRequired returns an environment variable as int or an error.
func GetEnvIntRequired(key string) (int, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return 0, ErrConfigRequired("GetEnvIntRequired", key)
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, ErrConfigParseError("GetEnvIntRequired", key, err)
	}
	return i, nil
}

// GetEnvInt64 returns an environment variable as an int64 or a default.
func GetEnvInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvFloat64 returns an environment variable as a float64 or a default.
func GetEnvFloat64(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// GetEnvBool returns an environment variable as a bool or a default.
func GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// GetEnvBoolRequired returns an environment variable as bool or an error.
func GetEnvBoolRequired(key string) (bool, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return false, ErrConfigRequired("GetEnvBoolRequired", key)
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, ErrConfigParseError("GetEnvBoolRequired", key, err)
	}
	return b, nil
}

// GetEnvDuration returns an environment variable as a duration or a default.
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

// GetEnvDurationRequired returns an environment variable as duration or an error.
func GetEnvDurationRequired(key string) (time.Duration, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return 0, ErrConfigRequired("GetEnvDurationRequired", key)
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, ErrConfigParseError("GetEnvDurationRequired", key, err)
	}
	return d, nil
}

// GetEnvSlice returns an environment variable as a string slice.
// Values are split by the separator (default comma).
func GetEnvSlice(key string, defaultValue []string, separator string) []string {
	if separator == "" {
		separator = ","
	}
	if value, exists := os.LookupEnv(key); exists && value != "" {
		parts := strings.Split(value, separator)
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

// ============================================================================
// .env File Loading
// ============================================================================

// LoadEnvFile loads environment variables from a .env file.
func LoadEnvFile(path string) error {
	return LoadEnvFileWithOptions(path, false)
}

// LoadEnvFileWithOptions loads a .env file with options.
// If override is true, existing environment variables will be overwritten.
func LoadEnvFileWithOptions(path string, override bool) error {
	const op = "config.LoadEnvFile"

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrConfigNotFound(op, path)
		}
		return ErrConfigFileError(op, path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		key, value, err := parseLine(line)
		if err != nil {
			return ErrConfigParseError(op, "line "+strconv.Itoa(lineNum), err)
		}

		// Set environment variable
		if override {
			os.Setenv(key, value)
		} else {
			// Only set if not already set
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ErrConfigFileError(op, path, err)
	}

	return nil
}

// LoadEnvFileIfExists loads a .env file if it exists, ignoring not found errors.
func LoadEnvFileIfExists(path string) error {
	err := LoadEnvFile(path)
	if err != nil && IsConfigNotFound(err) {
		return nil
	}
	return err
}

// parseLine parses a line from a .env file.
func parseLine(line string) (string, string, error) {
	// Handle export prefix
	line = strings.TrimPrefix(line, "export ")
	line = strings.TrimSpace(line)

	// Find the first '='
	idx := strings.Index(line, "=")
	if idx == -1 {
		return "", "", &parseError{msg: "invalid format, expected KEY=value"}
	}

	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])

	if key == "" {
		return "", "", &parseError{msg: "empty key"}
	}

	// Remove surrounding quotes from value
	value = unquote(value)

	return key, value, nil
}

// unquote removes surrounding quotes from a string.
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}

	// Check for matching quotes
	first := s[0]
	last := s[len(s)-1]

	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return s[1 : len(s)-1]
	}

	return s
}

// parseError represents a parsing error.
type parseError struct {
	msg string
}

func (e *parseError) Error() string {
	return e.msg
}

// ============================================================================
// Environment Variable Expansion
// ============================================================================

// ExpandEnv expands ${VAR} or $VAR in a string using environment variables.
func ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

// ExpandEnvMap expands environment variables in all values of a map.
func ExpandEnvMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = os.ExpandEnv(v)
	}
	return result
}

// ============================================================================
// Key Transformation
// ============================================================================

// ToEnvKey converts a field name to an environment variable key.
// e.g., "ServerPort" -> "SERVER_PORT"
func ToEnvKey(name string) string {
	var result strings.Builder
	result.Grow(len(name) + 5)

	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		if r >= 'a' && r <= 'z' {
			result.WriteByte(byte(r - 'a' + 'A'))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToEnvKeyWithPrefix converts a field name to an environment variable key with prefix.
// e.g., "APP", "ServerPort" -> "APP_SERVER_PORT"
func ToEnvKeyWithPrefix(prefix, name string) string {
	key := ToEnvKey(name)
	if prefix == "" {
		return key
	}
	return prefix + "_" + key
}
