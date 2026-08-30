package config

import (
	"os"
	"strconv"
	"time"
)

// Defaults for the safety limits the server enforces.
const (
	DefaultMaxBodyBytes  = 200 << 20 // max request body (upload) size
	DefaultMaxFileBytes  = 2 << 20   // max per-file size the reader will open
	DefaultMaxRepoBytes  = 200 << 20 // max workspace/clone/zip size
	DefaultMaxZipFiles   = 50_000    // max entries accepted from a zip upload
	DefaultRequestTimeout = 120 * time.Second
	DefaultCloneTimeout   = 120 * time.Second
)

type Config struct {
	Host             string
	Port             string
	RGBinary         string
	WatsonxAPIKey    string
	WatsonxProjectID string
	WatsonxModel     string
	MaxBodyBytes     int64
	MaxFileBytes     int64
	MaxRepoBytes     int64
	MaxZipFiles      int
	RequestTimeout   time.Duration
	CloneTimeout     time.Duration
}

func FromEnvironment() Config {
	return Config{
		Host:             env("HOST", "127.0.0.1"),
		Port:             env("PORT", "8080"),
		RGBinary:         env("RG_BINARY", "rg"),
		WatsonxAPIKey:    os.Getenv("WATSONX_API_KEY"),
		WatsonxProjectID: os.Getenv("WATSONX_PROJECT_ID"),
		WatsonxModel:     os.Getenv("WATSONX_MODEL"),
		MaxBodyBytes:     int64(envInt("MAX_BODY_BYTES", DefaultMaxBodyBytes)),
		MaxFileBytes:     int64(envInt("MAX_FILE_BYTES", DefaultMaxFileBytes)),
		MaxRepoBytes:     int64(envInt("MAX_REPO_BYTES", DefaultMaxRepoBytes)),
		MaxZipFiles:      envInt("MAX_ZIP_FILES", DefaultMaxZipFiles),
		RequestTimeout:   durationEnv("REQUEST_TIMEOUT_SECONDS", DefaultRequestTimeout),
		CloneTimeout:     durationEnv("CLONE_TIMEOUT_SECONDS", DefaultCloneTimeout),
	}
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
func durationEnv(key string, fallback time.Duration) time.Duration {
	seconds, err := strconv.Atoi(os.Getenv(key))
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}