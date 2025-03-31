package redisdb

import (
	"context"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/go-redis/redis/v8"
)

type RedisDB struct {
	*redis.Client
}

func NewRedisDB(conf *configs.Config) *RedisDB {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})

	// Проверяем соединение с Redis
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	return &RedisDB{client}
}
