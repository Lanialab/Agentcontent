package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DatabasePath      string
	OutlierMultiplier float64
	SeedOnEmpty       bool
	YouTubeAPIKey     string
	OpenAIKey         string
	OpenAIBaseURL     string
	OpenAIModel       string
	SheetsEnabled     bool
	SheetsID          string
}

func Load() Config {
	return Config{
		Port:              env("PORT", "8080"),
		DatabasePath:      env("DATABASE_PATH", "./data/agentcontent.db"),
		OutlierMultiplier: envFloat("OUTLIER_MULTIPLIER", 2.5),
		SeedOnEmpty:       envBool("SEED_ON_EMPTY", true),
		YouTubeAPIKey:     os.Getenv("YOUTUBE_API_KEY"),
		OpenAIKey:         os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:     env("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:       env("OPENAI_MODEL", "gpt-4o-mini"),
		SheetsEnabled:     envBool("SHEETS_ENABLED", false),
		SheetsID:          os.Getenv("SHEETS_SPREADSHEET_ID"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
