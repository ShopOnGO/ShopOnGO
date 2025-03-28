package passwordreset

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/redisdb"
)

type RedisResetRepository struct {
    redis *redisdb.RedisDB
}

func NewRedisResetRepository(r *redisdb.RedisDB) *RedisResetRepository {
    return &RedisResetRepository{redis: r}
}

func (r *RedisResetRepository) SaveToken(email, code string, expiresAt time.Time) error {
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        ttl = time.Minute // запас
    }
    key := r.key(email)
	logger.Info("🔑 Сохранение токена для email: " + email)
	err := r.redis.Set(context.Background(), key, code, ttl).Err()
    if err != nil {
        logger.Error("❌ ошибка при сохранении токена в Redis для email " + email + ": " + err.Error())
        return err
    }

    logger.Info("✅ Токен успешно сохранен для email: " + email)
    return nil
}

func (r *RedisResetRepository) GetToken(email string) (string, time.Time, error) {
    key := r.key(email)
    res, err := r.redis.Get(context.Background(), key).Result()
    if err == redis.Nil {
        return "", time.Time{}, fmt.Errorf("token not found")
    }
    if err != nil {
        return "", time.Time{}, err
    }

    ttl, err := r.redis.TTL(context.Background(), key).Result()
    if err != nil {
        return "", time.Time{}, err
    }

    expiresAt := time.Now().Add(ttl)

    return res, expiresAt, nil
}

func (r *RedisResetRepository) DeleteToken(email string) error {
    key := r.key(email)
    return r.redis.Del(context.Background(), key).Err()
}

func (r *RedisResetRepository) key(email string) string {
    return fmt.Sprintf("reset_token:%s", email)
}

func (r *RedisResetRepository) GetResetCodeCount(email string) (int, error) {
	key := r.requestCountKey(email)
	count, err := r.redis.Get(context.Background(), key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

func (r *RedisResetRepository) IncrementResetCodeCount(email string, ttl time.Duration) error {
	key := r.requestCountKey(email)
	count, err := r.redis.Incr(context.Background(), key).Result()
	if err != nil {
		return err
	}
	// Если ключ только создан (count == 1), установим время жизни
	if count == 1 {
		if err := r.redis.Expire(context.Background(), key, ttl).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisResetRepository) requestCountKey(email string) string {
	return fmt.Sprintf("reset_requests:%s", email)
}