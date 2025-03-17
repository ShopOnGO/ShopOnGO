package oauth2

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RefreshTokenRepository описывает методы для работы с refresh‑токенами.
type RefreshTokenRepository interface {
	GetUserIDByRefreshToken(refreshToken string) (string, error)
	StoreRefreshToken(userID, refreshToken string, expiresIn time.Duration) error
	DeleteRefreshToken(refreshToken string) error
}

// RedisRefreshTokenRepository реализует RefreshTokenRepository с помощью Redis.
type RedisRefreshTokenRepository struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisRefreshTokenRepository создаёт новый репозиторий.
func NewRedisRefreshTokenRepository(client *redis.Client) *RedisRefreshTokenRepository {
	return &RedisRefreshTokenRepository{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisRefreshTokenRepository) GetUserIDByRefreshToken(refreshToken string) (string, error) {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	userID, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("refresh token not found")
	}
	return userID, err
}

func (r *RedisRefreshTokenRepository) StoreRefreshToken(userID, refreshToken string, expiresIn time.Duration) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return r.client.Set(r.ctx, key, userID, expiresIn).Err()
}

func (r *RedisRefreshTokenRepository) DeleteRefreshToken(refreshToken string) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return r.client.Del(r.ctx, key).Err()
}
