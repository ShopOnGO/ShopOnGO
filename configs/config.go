package configs

import (
	"os"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2/oauth2manager"
	"github.com/joho/godotenv"
)

type Config struct {
	Db    DbConfig
	Auth  AuthConfig
	Redis RedisConfig
	OAuth OAuthConfig
}

type DbConfig struct {
	Dsn string
}

type AuthConfig struct {
	Secret string
	JWTTTL time.Duration
}

type RedisConfig struct {
	Addr            string
	Password        string
	DB              int
	RefreshTokenTTL time.Duration
}

type OAuthConfig struct { // Новая структура для OAuth2
	OAuth2Manager *oauth2manager.OAuth2ManagerImpl
	Secret        string
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
		jwtTTLStr = "1h"
	}
	jwtTTL, err := time.ParseDuration(jwtTTLStr)
	if err != nil {
		logger.Error("Invalid JWT_TTL, using default 1h", err.Error())
		jwtTTL = 1 * time.Hour
	}

	oauthManager := oauth2manager.NewOAuth2Manager(
		os.Getenv("REDIS_ADDRESS"),
		os.Getenv("REDIS_PASSWORD"),
		os.Getenv("SECRET"),  // Передаем секрет
		0,
	)

	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: AuthConfig{
			Secret: os.Getenv("SECRET"),
			JWTTTL: jwtTTL,
		},
		Redis: RedisConfig{
			Addr:            os.Getenv("REDIS_ADDRESS"),
			Password:        os.Getenv("REDIS_PASSWORD"),
			DB:              0,
			RefreshTokenTTL: refreshTTL,
		},
		OAuth: OAuthConfig{
			OAuth2Manager: oauthManager,
		},
	}
}
