// Package env provides helper functions for reading configuration values and environment variables safely with fallback defaults.
package env

import (
	"os"
	"strconv"
)

// GetString retrieves a string environment variable by key, returning the provided fallback value if the key is unset.
func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

// GetInt retrieves an integer environment variable by key, returning the provided fallback value if the key is unset or fails integer parsing.
func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valAsInt
}

// GetBoolean retrieves a boolean environment variable by key, returning the provided fallback value if the key is unset or fails boolean parsing.
func GetBoolean(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsBoolean, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}

	return valAsBoolean
}
