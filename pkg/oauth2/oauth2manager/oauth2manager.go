package oauth2manager

import (
	"context"
	"errors"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	_ "github.com/go-oauth2/oauth2/v4/store"
	redisStore "github.com/go-oauth2/redis/v4"
	"github.com/go-redis/redis/v8"
)

// OAuth2Manager описывает интерфейс для генерации/обновления токенов.
// Здесь можно задать любые методы, которые вам нужны.
type OAuth2Manager interface {
	GenerateTokens(data interface{}) (accessToken, refreshToken string, err error)
	RefreshTokens(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
}

type OAuth2ManagerImpl struct {
	Manager *manage.Manager
}

// NewOAuth2Manager создает менеджер OAuth2 с использованием Redis для хранения токенов.
func NewOAuth2Manager(redisAddr, redisPassword string, redisDB int) *OAuth2ManagerImpl {
	// Создаем менеджер с настройками по умолчанию
	m := manage.NewDefaultManager()

	// Инициализируем redis-клиент
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Создаем redis-хранилище
	// Вторым параметром ("oauth2:") указываем префикс для ключей в Redis (необязательно).
	tokenStore := redisStore.NewRedisStoreWithCli(rdb, "oauth2:")

	// Привязываем хранилище токенов к менеджеру
	m.MapTokenStorage(tokenStore)

	// Настраиваем конфиг (время жизни токенов, генерация refresh-токена и т.д.)
	cfg := &manage.Config{
		AccessTokenExp:    time.Hour,           // access-токен живет 1 час
		RefreshTokenExp:   30 * 24 * time.Hour, // refresh-токен живет 30 дней
		IsGenerateRefresh: true,                // генерировать refresh-токен
	}

	// Выберите нужный flow. Например, Password Credentials:
	m.SetPasswordTokenCfg(cfg)
	// Для Client Credentials flow:
	// m.SetClientTokenCfg(cfg)
	// Для Authorization Code flow:
	// m.SetAuthorizeCodeTokenCfg(cfg)

	// Генератор access-токена (по умолчанию Bearer Token).
	// Можно заменить на генерацию JWT, используя generates.NewJWTAccessGenerate(...)
	m.MapAccessGenerate(generates.NewAccessGenerate())

	return &OAuth2ManagerImpl{
		Manager: m,
	}
}

// GenerateTokens генерирует access и refresh токены, используя Password Flow.
func (o *OAuth2ManagerImpl) GenerateTokens(data interface{}) (string, string, error) {
	ctx := context.Background()

	// Пример: здесь предполагается, что data содержит userID.
	userID, ok := data.(string) // Ожидаем, что передаётся userID (например, email)
	if !ok {
		return "", "", errors.New("invalid token data type")
	}

	// Создаём запрос на генерацию токена
	tgr := &oauth2.TokenGenerateRequest{
		ClientID: "default", // Укажите client_id
		UserID:   userID,    // ID пользователя
		Scope:    "",
	}

	// Генерируем токен
	ti, err := o.Manager.GenerateAccessToken(ctx, oauth2.PasswordCredentials, tgr)
	if err != nil {
		return "", "", err
	}

	accessToken := ti.GetAccess()
	refreshToken := ti.GetRefresh()
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
