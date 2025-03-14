package oauth2manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/jwt"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// OAuth2Manager описывает интерфейс для генерации/обновления токенов.
// Здесь можно задать любые методы, которые вам нужны.
type OAuth2Manager interface {
	GenerateTokens(data interface{}) (accessToken, refreshToken string, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
}

type OAuth2ManagerImpl struct {
	Manager     *manage.Manager
	RedisClient *redis.Client
	Secret      string
}

// NewOAuth2Manager создает менеджер OAuth2 с использованием Redis для хранения токенов.
func NewOAuth2Manager(redisAddr, redisPassword string, secret string, redisDB int) *OAuth2ManagerImpl {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	tokenStore, err := store.NewMemoryTokenStore()
	if err != nil {
		logger.Errorf("failed to create token store: %v", err)
	}
	manager.MustTokenStorage(tokenStore, err)

	// ⚠️ Добавляем пустой clientStore, но не используем ClientID
	clientStore := store.NewClientStore()
	clientStore.Set("default", &models.Client{
		ID:     "default",
		Secret: "", // Можно оставить пустым
		Domain: "", // Необязательно
	})

	manager.MapClientStorage(clientStore)

	return &OAuth2ManagerImpl{
		Manager:     manager,
		RedisClient: client,
		Secret:      secret,
	}
}

// GenerateTokens генерирует access и refresh токены, используя Password Flow.
func (o *OAuth2ManagerImpl) GenerateTokens(data interface{}) (string, string, error) {
	// Пример: здесь предполагается, что data содержит userID.
	userID, ok := data.(string) // Ожидаем, что передаётся userID (например, email)
	if !ok {
		return "", "", errors.New("invalid token data type")
	}

    accessToken, err := jwt.NewJWT(o.Secret).Create(jwt.JWTData{Email: userID}, 15*time.Minute)
    if err != nil {
        return "", "", err
    }

	logger.Info("Генерация токена для пользователя:", userID)
	// Создаём запрос на генерацию токена
	tgr := &oauth2.TokenGenerateRequest{
		ClientID: "default", // Должен совпадать с clientStore
		UserID:   userID,
	}

	// Генерируем токен
	ti, err := o.Manager.GenerateAccessToken(ctx, oauth2.PasswordCredentials, tgr)
	if err != nil {
		logger.Error("Ошибка при генерации токена:", err)
		return "", "", err
	}

	log.Println("Токен сгенерирован:", ti.GetRefresh())

	refreshToken := ti.GetRefresh()

	// Сохраняем refresh_token в Redis
	err = o.StoreRefreshToken(userID, refreshToken, 30*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens обновляет access-токен с использованием refresh-токена.
func (o *OAuth2ManagerImpl) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	// Загружаем информацию о токене по refresh-токену
	tokenInfo, err := o.Manager.LoadRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	// Создаем запрос на обновление токена
	tgr := &oauth2.TokenGenerateRequest{
		ClientID: tokenInfo.GetClientID(),
		UserID:   tokenInfo.GetUserID(),
		Scope:    tokenInfo.GetScope(),
		Refresh:  refreshToken,
	}

	// Генерируем новый access-токен
	ti, err := o.Manager.GenerateAccessToken(ctx, oauth2.Refreshing, tgr)
	if err != nil {
		return "", "", err
	}

	accessToken := ti.GetAccess()
	newRefreshToken := ti.GetRefresh()
	return accessToken, newRefreshToken, nil
}

func (o *OAuth2ManagerImpl) GetUserIDByRefreshToken(refreshToken string) (string, error) {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	userID, err := o.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("refresh token not found")
	}
	return userID, err
}

// Сохранение refresh_token -> user_id в Redis
func (o *OAuth2ManagerImpl) StoreRefreshToken(userID, refreshToken string, expiresIn time.Duration) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return o.RedisClient.Set(ctx, key, userID, expiresIn).Err()
}

// Удаление refresh_token (при разлогине)
func (o *OAuth2ManagerImpl) DeleteRefreshToken(refreshToken string) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return o.RedisClient.Del(ctx, key).Err()
}
