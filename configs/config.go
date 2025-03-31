package configs

import (
	"os"
	"strconv"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	Db    DbConfig
	Redis RedisConfig
	OAuth OAuthConfig
	Google GoogleConfig
	Code CodeConfig
	SMTP SMTPConfig
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

type OAuthConfig struct {
	Secret        string
	JWTTTL time.Duration
}

type CodeConfig struct {
	CodeTTL time.Duration
	MaxRequests int
	RateLimitTTL time.Duration
}

type SMTPConfig struct {
	Name string
	From string
	Pass string
	Host string
	Port int
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

	codeTTLStr := os.Getenv("CODE_TTL")
	if codeTTLStr == "" {
		codeTTLStr = "10m"
	}
	codeTTL, err := time.ParseDuration(codeTTLStr)
	if err != nil {
		logger.Error("Invalid CODE_TTL, using default 10m", err.Error())
		codeTTL = 10 * time.Minute
	}
	maxRequestsStr := os.Getenv("CODE_MAX_REQUESTS")
	maxRequests := 5
	if maxRequestsStr != "" {
		if val, err := strconv.Atoi(maxRequestsStr); err == nil {
			maxRequests = val
		} else {
			logger.Error("Invalid CODE_MAX_REQUESTS, using default 5", err.Error())
		}
	}
	rateLimitTTLStr := os.Getenv("CODE_RATE_LIMIT_TTL")
	rateLimitTTL := 24 * time.Hour
	if rateLimitTTLStr != "" {
		if val, err := time.ParseDuration(rateLimitTTLStr); err == nil {
			rateLimitTTL = val
		} else {
			logger.Error("Invalid CODE_RATE_LIMIT_TTL, using default 24h", err.Error())
		}
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
		Code: CodeConfig{
            CodeTTL: codeTTL,
			MaxRequests: maxRequests,
			RateLimitTTL: rateLimitTTL,
        },
		SMTP: SMTPConfig{
			Name: os.Getenv("SMTP_NAME"),
			From: os.Getenv("SMTP_FROM"),
            Pass: os.Getenv("SMTP_PASS"),
            Host: os.Getenv("SMTP_HOST"),
            Port: 587, // TLS
		},
	}
}
