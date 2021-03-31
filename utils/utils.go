package utils

import (
	"os"
	"strconv"
)

var PostgresConn = "postgres://postgres:postgres@localhost:5433/postgres"

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
func GetEnvAsInt(name string, defaultVal int64) int64 {
	valueStr := GetEnv(name, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}
func GetEnvAsBool(name string, defaultVal bool) bool {
	valStr := GetEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}
