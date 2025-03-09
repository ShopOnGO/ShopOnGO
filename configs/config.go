package configs

import (
	"os"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	Db    DbConfig
	Auth  AuthConfig
	Redis RedisConfig
}
type DbConfig struct {
	Dsn string
}
type AuthConfig struct {
	Secret string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func LoadConfig() *Config {
	err := godotenv.Load() //loading from .env
	if err != nil {
		logger.Error("Error loading .env file, using default config", err.Error())
	}
	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: AuthConfig{
			Secret: os.Getenv("SECRET"),
		},
		Redis: RedisConfig{
			Addr:     os.Getenv("REDIS_ADDRESS"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		},
	}
}
