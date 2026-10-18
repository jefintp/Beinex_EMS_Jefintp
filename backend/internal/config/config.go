package config

import (
	"fmt"
	"net/url"
	"os"
)

// Config holds runtime configuration read from the environment.
type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	CORSOrigin string
}

// Load reads configuration from environment variables with sane defaults.
func Load() Config {
	return Config{
		Port:       getenv("PORT", "8080"),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBName:     getenv("DB_NAME", "employee_db"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"), // use "require" on most hosted databases
		CORSOrigin: getenv("CORS_ORIGIN", "http://localhost:4200"),
	}
}

// DSN builds the PostgreSQL connection URL (credentials are URL-escaped).
func (c Config) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.DBUser, c.DBPassword),
		Host:     fmt.Sprintf("%s:%s", c.DBHost, c.DBPort),
		Path:     "/" + c.DBName,
		RawQuery: url.Values{"sslmode": []string{c.DBSSLMode}}.Encode(),
	}
	return u.String()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
