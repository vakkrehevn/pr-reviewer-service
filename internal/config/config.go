package config

import "os"

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	Host     = getEnv("DB_HOST", "localhost")
	Port     = 5432
	User     = getEnv("DB_USER", "postgres")
	Password = getEnv("DB_PASSWORD", "password")
	DBName   = getEnv("DB_NAME", "pr_reviewer")
)
