// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      []byte
	JWTTTL         time.Duration
	FrontendOrigin string
	// Production toggles Secure + SameSite=None on the auth cookie so it can be
	// sent from a frontend hosted on a different domain than the API.
	Production bool

	// Optional demo admin: when both are set, an admin account is seeded on
	// startup so reviewers can try the "view all tasks" admin feature without
	// running SQL. Leave unset in real deployments.
	SeedAdminEmail    string
	SeedAdminPassword string
}

// Load reads configuration from the environment, returning an error when a
// required value is missing so the server fails fast instead of half-booting.
func Load() (*Config, error) {
	c := &Config{
		Port:              getenv("PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		FrontendOrigin:    getenv("FRONTEND_ORIGIN", "http://localhost:3000"),
		Production:        getenv("APP_ENV", "development") == "production",
		SeedAdminEmail:    os.Getenv("SEED_ADMIN_EMAIL"),
		SeedAdminPassword: os.Getenv("SEED_ADMIN_PASSWORD"),
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET is required and must be at least 32 characters")
	}
	c.JWTSecret = []byte(secret)

	ttlHours, _ := strconv.Atoi(getenv("JWT_TTL_HOURS", "168")) // default 7 days
	c.JWTTTL = time.Duration(ttlHours) * time.Hour

	return c, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
