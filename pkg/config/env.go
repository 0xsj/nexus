// Package config provides utilities for loading configuration from environment variables.
// This package follows a decentralized configuration approach where each domain
// module loads its own configuration using these primitives.
package config

import (
	"os"
	"strings"
)

// GetEnv retrieves an environment variable with an optional default value.
// If the environment variable is not set or is empty, the default value is returned.
//
// Example:
//
//	port := config.GetEnv("PORT", "8080")
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvRequired retrieves a required environment variable.
// Returns the value and a boolean indicating whether the variable was found.
// The caller is responsible for handling the missing variable case.
//
// Example:
//
//	apiKey, ok := config.GetEnvRequired("API_KEY")
//	if !ok {
//	    return fmt.Errorf("API_KEY is required")
//	}
func GetEnvRequired(key string) (string, bool) {
	value := os.Getenv(key)
	return value, value != ""
}

// GetEnvWithPrefix retrieves an environment variable with a prefix prepended.
// This is useful for domain-specific configuration namespacing.
//
// Example:
//
//	// Looks for "PROFILE_MAX_LINKS"
//	maxLinks := config.GetEnvWithPrefix("PROFILE_", "MAX_LINKS")
func GetEnvWithPrefix(prefix, key string) string {
	return os.Getenv(prefix + key)
}

// HasEnv checks if an environment variable exists, even if it's set to an empty string.
// This is different from checking if a value is non-empty.
//
// Example:
//
//	if config.HasEnv("FEATURE_FLAG") {
//	    // Variable is explicitly set (even if empty)
//	}
func HasEnv(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// GetEnvMap retrieves all environment variables with a given prefix.
// Returns a map where keys have the prefix removed.
//
// Example:
//
//	// With env vars: PROFILE_MAX_LINKS=10, PROFILE_CACHE_TTL=3600
//	config := config.GetEnvMap("PROFILE_")
//	// Returns: {"MAX_LINKS": "10", "CACHE_TTL": "3600"}
func GetEnvMap(prefix string) map[string]string {
	result := make(map[string]string)

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key, value := pair[0], pair[1]
		if strings.HasPrefix(key, prefix) {
			// Remove prefix from key
			cleanKey := strings.TrimPrefix(key, prefix)
			result[cleanKey] = value
		}
	}

	return result
}
