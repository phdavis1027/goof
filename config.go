package main

import (
	"os"
)

type Config struct {
	MeilisearchURL string
	MeilisearchKey string
	IndexName      string
}

func LoadConfig() Config {
	return Config{
		MeilisearchURL: getEnvOrDefault("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchKey: getEnvOrDefault("MEILISEARCH_KEY", ""),
		IndexName:      getEnvOrDefault("MEILISEARCH_INDEX", "error_reports"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}