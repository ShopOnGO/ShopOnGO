package configs

import (
	"os"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	Db    DbConfig
	Redis RedisConfig
	OAuth OAuthConfig
	Google GoogleConfig
}

type DbConfig struct {
	Dsn string
}

type RedisConfig struct {
	Addr            string
	Password        string
	DB              int
	RefreshTokenTTL time.Duration
}

type OAuthConfig struct { // Новая структура для OAuth2
	Secret        string
	JWTTTL time.Duration
}

type GoogleConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}

func LoadConfig() *Config {
	err := godotenv.Load() //loading from .env
	if err != nil {
		logger.Error("Error loading .env file, using default config", err.Error())
	}

	ttlStr := os.Getenv("REFRESH_TOKEN_TTL")
	if ttlStr == "" {
		ttlStr = "24h"
	}
	refreshTTL, err := time.ParseDuration(ttlStr)
	if err != nil {
		logger.Error("Invalid REFRESH_TOKEN_TTL, using default 24h", err.Error())
		refreshTTL = 24 * time.Hour
	}

	jwtTTLStr := os.Getenv("JWT_TTL")
	if jwtTTLStr == "" {
		jwtTTLStr = "15m"
	}
	jwtTTL, err := time.ParseDuration(jwtTTLStr)
	if err != nil {
		logger.Error("Invalid JWT_TTL, using default 1h", err.Error())
		jwtTTL = 15 * time.Minute
	}

	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Redis: RedisConfig{
			Addr:            os.Getenv("REDIS_ADDRESS"),
			// Password:        os.Getenv("REDIS_PASSWORD"),
			Password:        "",
			DB:              0,
			RefreshTokenTTL: refreshTTL,
		},
		OAuth: OAuthConfig{
			Secret: os.Getenv("SECRET"),
			JWTTTL: jwtTTL,
		},
		Google: GoogleConfig{
			ClientID: os.Getenv("CLIENT_ID"),
			ClientSecret: os.Getenv("CLIENT_SECRET"),
			RedirectURL: os.Getenv("REDIRECT_URL"),
		},
	}
}
