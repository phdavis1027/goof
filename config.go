package main

import (
	"log"
	"os"
)

type Config struct {
	MeilisearchURL string
	MeilisearchKey string
	IndexName      string
}

func LoadConfig() Config {
	config := Config{
		MeilisearchURL: getEnvOrDefault("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchKey: getEnvOrDefault("MEILISEARCH_KEY", "aSampleMasterKey"),
		IndexName:      getEnvOrDefault("MEILISEARCH_INDEX", "error_reports"),
	}
	
	logToFile("DEBUG: Config loaded - URL: %s, Key: '%s' (len=%d), Index: %s\n", 
		config.MeilisearchURL, config.MeilisearchKey, len(config.MeilisearchKey), config.IndexName)
	
	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	logToFile("DEBUG: getEnvOrDefault(%s, %s) - env value: '%s' (len=%d)\n", 
		key, defaultValue, value, len(value))
	
	if value != "" {
		logToFile("DEBUG: Using env value: '%s'\n", value)
		return value
	}
	
	logToFile("DEBUG: Using default value: '%s'\n", defaultValue)
	return defaultValue
}

func logToFile(format string, args ...interface{}) {
	file, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()
	
	logger := log.New(file, "", log.LstdFlags)
	logger.Printf(format, args...)
}
