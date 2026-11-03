package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	APIPort          string
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	CookieSecure     bool
	CORSOrigins      []string
	SeedAdminEmail    string
	SeedAdminPassword string
	StorageDriver     string
	LocalStoragePath  string
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKey       string
	S3SecretKey       string
	OpenAIAPIKey      string
	OpenAIModel         string
}

func Load() Config {
	accessTTL, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 168 * time.Hour
	}
	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://moneymanager:changeme@localhost:5432/moneymanager?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		APIPort:           getEnv("API_PORT", "8080"),
		JWTAccessSecret:   getEnv("JWT_ACCESS_SECRET", "dev-access-secret-minimum-32-characters"),
		JWTRefreshSecret:  getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret-minimum-32-characters"),
		JWTAccessTTL:      accessTTL,
		JWTRefreshTTL:     refreshTTL,
		CookieSecure:      getEnv("COOKIE_SECURE", "false") == "true",
		CORSOrigins:       strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
		SeedAdminEmail:    os.Getenv("SEED_ADMIN_EMAIL"),
		SeedAdminPassword: os.Getenv("SEED_ADMIN_PASSWORD"),
		StorageDriver:     getEnv("STORAGE_DRIVER", "local"),
		LocalStoragePath:  getEnv("LOCAL_STORAGE_PATH", "./uploads"),
		S3Endpoint:        getEnv("S3_ENDPOINT", "http://minio:9000"),
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		S3Bucket:          getEnv("S3_BUCKET", "money-manager"),
		S3AccessKey:       getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:       getEnv("S3_SECRET_KEY", "minioadmin"),
		OpenAIAPIKey:      os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:       getEnv("OPENAI_MODEL", "gpt-4o-mini"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
