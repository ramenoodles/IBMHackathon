package config

import (
	"os"
	"strconv"
)

type Config struct {
	Host             string
	Port             string
	RGBinary         string
	WatsonxAPIKey    string
	WatsonxProjectID string
	WatsonxModel     string
	MaxBodyBytes     int64
}

func FromEnvironment() Config {
	return Config{Host: env("HOST", "127.0.0.1"), Port: env("PORT", "8080"), RGBinary: env("RG_BINARY", "rg"), WatsonxAPIKey: os.Getenv("WATSONX_API_KEY"), WatsonxProjectID: os.Getenv("WATSONX_PROJECT_ID"), WatsonxModel: os.Getenv("WATSONX_MODEL"), MaxBodyBytes: int64(envInt("MAX_BODY_BYTES", 200<<20))}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
