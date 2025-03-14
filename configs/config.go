package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Db   DbConfig
	Auth AuthConfig
}
type DbConfig struct {
	Dsn string
}
type AuthConfig struct {
	Secret string
}

func LoadConfig() *Config {
	err := godotenv.Load() //loading from .env
	if err != nil {
		log.Println("Error loading .env file, using dwfault config")
		log.Println(err.Error())
	}
	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: AuthConfig{
			Secret: os.Getenv("SECRET"),
		},
	}
}
