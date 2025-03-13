package cache

import (
	"context"
	"log"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	*redis.Client
}

func NewRedis(conf *configs.Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,     // "localhost:6379"
		Password: conf.Redis.Password, // Если пароль не нужен, оставляем пустым
		DB:       conf.Redis.DB,       // Номер БД Redis (по умолчанию 0)
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}

	return &Cache{client}

}
