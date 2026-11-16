package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr             string
	DatabaseURL          string
	OpenAIAPIKey         string
	EmbeddingDimensions  int
}

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dimensions := 1536
	if value := os.Getenv("EMBEDDING_DIMENSIONS"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			dimensions = parsed
		}
	}

	return Config{
		HTTPAddr:            addr,
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		EmbeddingDimensions: dimensions,
	}
}
