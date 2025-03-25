package oauth2

import (
	"context"
	"fmt"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/jwt"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
)

// OAuth2Service определяет бизнес-методы для работы с токенами.
type OAuth2Service interface {
	GenerateTokens(userID, role string) (accessToken, refreshToken string, err error)
	RefreshTokens(refreshToken string) (accessToken, newRefreshToken string, err error)
	Logout(refreshToken string) error
	// UserRole(email string) (string, error)
}

// oauth2ServiceImpl – реализация OAuth2Service.
type oauth2ServiceImpl struct {
	manager    *manage.Manager
	repo       RefreshTokenRepository
	secret     string
	jwtTTL     time.Duration
	refreshTTL time.Duration
	ctx        context.Context
}

// NewOAuth2Service создаёт новый сервис, используя конфигурацию и репозиторий.
func NewOAuth2Service(config *configs.Config, repo RefreshTokenRepository) OAuth2Service {
	// Инициализация менеджера OAuth2
	mgr := manage.NewDefaultManager()
	mgr.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	// Используем in-memory хранилище токенов (при необходимости можно заменить на другое)
	tokenStore, err := store.NewMemoryTokenStore()
	if err != nil {
		logger.Errorf("failed to create token store: %v", err)
	}
	mgr.MustTokenStorage(tokenStore, err)

	// Создаём хранилище клиентов с дефолтным клиентом
	clientStore := store.NewClientStore()
	clientStore.Set("default", &models.Client{
		ID:     "default",
		Secret: "", // Можно оставить пустым
		Domain: "",
	})
	mgr.MapClientStorage(clientStore)

	return &oauth2ServiceImpl{
		manager:    mgr,
		repo:       repo,
		secret:     config.OAuth.Secret,
		jwtTTL:     config.OAuth.JWTTTL,
		refreshTTL: config.Redis.RefreshTokenTTL,
		ctx:        context.Background(),
	}
}

// GenerateTokens генерирует новый access‑и refresh‑токены.
func (s *oauth2ServiceImpl) GenerateTokens(userID, role string) (string, string, error) {
	// Генерация access‑токена с использованием JWT
	accessToken, err := jwt.NewJWT(s.secret).Create(jwt.JWTData{Email: userID, Role: role}, s.jwtTTL)
	if err != nil {
		return "", "", err
	}

	logger.Info("Генерация токенов для пользователя:", userID)

	// Формируем запрос для генерации токена через OAuth2 менеджер
	tgr := &oauth2.TokenGenerateRequest{
		ClientID: "default",
		UserID:   userID,
	}

	ti, err := s.manager.GenerateAccessToken(s.ctx, oauth2.PasswordCredentials, tgr)
	if err != nil {
		logger.Error("Ошибка при генерации токена:", err)
		return "", "", err
	}

	refreshToken := ti.GetRefresh()

	data := &RefreshTokenData{
		UserID: userID,
		Role:   role,
	}
	// Сохраняем refresh‑токен через репозиторий
	err = s.repo.StoreRefreshToken(data, refreshToken, s.refreshTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens обновляет access‑токен с использованием переданного refresh‑токена.
func (s *oauth2ServiceImpl) RefreshTokens(refreshToken string) (string, string, error) {
	// Получаем userID из хранилища refresh‑токена
	data, err := s.repo.GetRefreshTokenData(refreshToken)
	if err != nil {
		return "", "", ErrInvalidOrExpiredRefreshToken
	}

	newAccessToken, newRefreshToken, err := s.GenerateTokens(data.UserID, data.Role)
	if err != nil {
		return "", "", ErrFailedToCreateNewTokens
	}
	
	// Обновляем refresh‑токен в хранилище
	err = s.repo.StoreRefreshToken(data, newRefreshToken, s.refreshTTL)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrFailedToStoreRefreshToken, err)
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout удаляет refresh-токен, вызывая метод репозитория.
func (s *oauth2ServiceImpl) Logout(refreshToken string) error {
	return s.repo.DeleteRefreshToken(refreshToken)
}