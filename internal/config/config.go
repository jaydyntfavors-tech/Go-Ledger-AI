package config

import (
  "log"
  "os"
)

type Config struct {
  Port         string
  DatabaseURL  string 
  GeminiAPIKEY string
}

func LoadConfig() *Config {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8000"
  }

  dbURL := od.Getenv("DATABASE_URL")
  if dbURL == "" {
    log.println("WARNING: DATABASE_URL is not set")
  }

  geminiKey := od.getenv("GEMINI_API_KEY")
  if geminiKey == "" {
    log.println("WARNING: GEMINI_API_KEY is not set")
  }

  return &Config{
    Port:                port,
    DatabaseURL:         dbURL,
    GeminiAPIKey:        geminiKey,
  }
}
