package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RefreshTokenData struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// RefreshTokenRepository описывает методы для работы с refresh‑токенами.
type RefreshTokenRepository interface {
	GetRefreshTokenData(refreshToken string) (*RefreshTokenData, error)
	StoreRefreshToken(data *RefreshTokenData, refreshToken string, expiresIn time.Duration) error
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

func (r *RedisRefreshTokenRepository) StoreRefreshToken(data *RefreshTokenData, refreshToken string, expiresIn time.Duration) error {
	oldKey := fmt.Sprintf("refresh:user:%s", data.UserID)
	newKey := fmt.Sprintf("refresh:%s", refreshToken)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	oldToken, err := r.client.Get(r.ctx, oldKey).Result()
	if err == redis.Nil {
		oldToken = ""
	} else if err != nil {
		return err
	}

	_, err = r.client.TxPipelined(r.ctx, func(pipe redis.Pipeliner) error {
		if oldToken != "" {
			pipe.Del(r.ctx, fmt.Sprintf("refresh:%s", oldToken))
		}
		pipe.Set(r.ctx, newKey, jsonData, expiresIn)
		pipe.Set(r.ctx, oldKey, refreshToken, expiresIn)

		return nil
	})
	return err
}

func (r *RedisRefreshTokenRepository) GetRefreshTokenData(refreshToken string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	jsonData, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, err
	}
	var data RefreshTokenData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *RedisRefreshTokenRepository) DeleteRefreshToken(refreshToken string) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return r.client.Del(r.ctx, key).Err()
}
