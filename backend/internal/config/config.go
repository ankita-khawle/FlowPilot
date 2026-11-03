package config

import "os"

// Config contains runtime settings for the FlowPilot API.
type Config struct {
	HTTPAddr string
}

// Load reads configuration from environment variables and applies safe defaults.
func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{HTTPAddr: addr}
}
