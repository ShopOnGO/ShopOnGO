package oauth2_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2"
)

type RefreshTokenData struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// fakeRefreshTokenRepo – фиктивная реализация интерфейса RefreshTokenRepository.
type fakeRefreshTokenRepo struct {
	storeFunc  func(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error
	getFunc    func(refreshToken string) (*oauth2.RefreshTokenData, error)
	deleteFunc func(refreshToken string) error
}

func (f *fakeRefreshTokenRepo) StoreRefreshToken(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error {
	if f.storeFunc != nil {
		return f.storeFunc(data, refreshToken, expiresIn)
	}
	return nil
}

func (f *fakeRefreshTokenRepo) GetRefreshTokenData(refreshToken string) (*oauth2.RefreshTokenData, error) {
	if f.getFunc != nil {
		return f.getFunc(refreshToken)
	}
	return nil, errors.New("token not found")
}

func (f *fakeRefreshTokenRepo) DeleteRefreshToken(refreshToken string) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(refreshToken)
	}
	return nil
}

type OAuth2Service interface {
	GenerateTokens(userID, role string) (string, string, error)
	RefreshTokens(refreshToken string) (string, string, error)
	Logout(refreshToken string) error
}

// Для простоты тестирования можно создать минимальную конфигурацию и использовать in-memory хранилище.
func getTestConfig() *configs.Config {
	return &configs.Config{
		OAuth: configs.OAuthConfig{
			Secret: "testsecret",
			JWTTTL: time.Hour,
		},
		Redis: configs.RedisConfig{
			RefreshTokenTTL: 24 * time.Hour,
		},
	}
}


// TestGenerateTokens_Success проверяет успешную генерацию токенов.
func TestGenerateTokens_Success(t *testing.T) {
	conf := getTestConfig()

	// Создаем фиктивный репозиторий, который всегда возвращает nil.
	fakeRepo := &fakeRefreshTokenRepo{
		storeFunc: func(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error {
			return nil
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	accessToken, refreshToken, err := service.GenerateTokens("user@example.com", "admin")
	if err != nil {
		t.Fatalf("ожидалась успешная генерация токенов, получена ошибка: %v", err)
	}
	if accessToken == "" {
		t.Error("ожидался непустой accessToken, получено пустое значение")
	}
	if refreshToken == "" {
		t.Error("ожидался непустой refreshToken, получено пустое значение")
	}

	// Можно дополнительно проверить, что accessToken является корректной JWT-строкой,
	// а refreshToken не пуст.
	t.Logf("AccessToken: %s", accessToken)
	t.Logf("RefreshToken: %s", refreshToken)
}
// TestGenerateTokens_RepoStoreError проверяет, что при ошибке сохранения refresh-токена возвращается ошибка.
func TestGenerateTokens_RepoStoreError(t *testing.T) {
	conf := getTestConfig()

	expectedErr := errors.New("store error")
	fakeRepo := &fakeRefreshTokenRepo{
		storeFunc: func(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error {
			return expectedErr
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	_, _, err := service.GenerateTokens("user@example.com", "admin")
	if err == nil {
		t.Error("ожидалась ошибка при сохранении refresh-токена, получено nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("ожидалась ошибка %v, получена %v", expectedErr, err)
	}
}
// Дополнительный тест: можно проверить, что данные RefreshTokenData корректно сериализуются
// и сохраняются в репозитории. Для этого можно расширить фиктивный репозиторий, сохраняя переданные данные.
func TestGenerateTokens_StoredData(t *testing.T) {
	conf := getTestConfig()

	var storedData *oauth2.RefreshTokenData
	var storedToken string
	var storedTTL time.Duration

	fakeRepo := &fakeRefreshTokenRepo{
		storeFunc: func(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error {
			storedData = data
			storedToken = refreshToken
			storedTTL = expiresIn
			// Имитация успешного сохранения
			return nil
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	userID := "user@example.com"
	role := "admin"
	_, refreshToken, err := service.GenerateTokens(userID, role)
	if err != nil {
		t.Fatalf("ошибка генерации токенов: %v", err)
	}

	// Проверяем, что в репозитории были переданы правильные данные
	if storedData == nil {
		t.Fatal("данные не были сохранены в репозитории")
	}
	if storedData.UserID != userID || storedData.Role != role {
		t.Errorf("ожидались данные UserID=%s, Role=%s, получены UserID=%s, Role=%s", userID, role, storedData.UserID, storedData.Role)
	}
	if storedToken != refreshToken {
		t.Errorf("ожидался refreshToken=%s, получен %s", refreshToken, storedToken)
	}
	if storedTTL != conf.Redis.RefreshTokenTTL {
		t.Errorf("ожидалось время жизни %v, получено %v", conf.Redis.RefreshTokenTTL, storedTTL)
	}

	// Дополнительно можно проверить, что JSON сериализация работает корректно:
	jsonData, err := json.Marshal(storedData)
	if err != nil {
		t.Errorf("ошибка сериализации данных: %v", err)
	}
	t.Logf("Сохраненные данные в JSON: %s", string(jsonData))
}


// TestRefreshTokens_Success проверяет успешное обновление токенов.
func TestRefreshTokens_Success(t *testing.T) {
	conf := getTestConfig()

	fakeRepo := &fakeRefreshTokenRepo{
		storeFunc: func(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error {
			return nil
		},
		getFunc: func(refreshToken string) (*oauth2.RefreshTokenData, error) {
			return &oauth2.RefreshTokenData{UserID: "user@example.com", Role: "admin"}, nil
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	newAccessToken, newRefreshToken, err := service.RefreshTokens("valid_refresh_token")
	if err != nil {
		t.Fatalf("ожидалось успешное обновление токенов, получена ошибка: %v", err)
	}
	if newAccessToken == "" || newRefreshToken == "" {
		t.Error("ожидались непустые токены, получены пустые значения")
	}
}

// TestRefreshTokens_InvalidToken проверяет ошибку при недействительном refresh-токене.
func TestRefreshTokens_InvalidToken(t *testing.T) {
	conf := getTestConfig()

	fakeRepo := &fakeRefreshTokenRepo{
		getFunc: func(refreshToken string) (*oauth2.RefreshTokenData, error) {
			return nil, errors.New("invalid or expired refresh token")
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	_, _, err := service.RefreshTokens("invalid_refresh_token")
	if err == nil || err.Error() != "invalid or expired refresh token" {
		t.Errorf("ожидалась ошибка 'invalid or expired refresh token', получено: %v", err)
	}
}

// TestLogout_Success проверяет успешное удаление refresh-токена.
func TestLogout_Success(t *testing.T) {
	conf := getTestConfig()

	fakeRepo := &fakeRefreshTokenRepo{
		deleteFunc: func(refreshToken string) error {
			return nil
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	err := service.Logout("valid_refresh_token")
	if err != nil {
		t.Fatalf("ожидалось успешное удаление токена, получена ошибка: %v", err)
	}
}

// TestLogout_Failure проверяет ошибку при удалении refresh-токена.
func TestLogout_Failure(t *testing.T) {
	conf := getTestConfig()

	expectedErr := errors.New("failed to delete refresh token")

	fakeRepo := &fakeRefreshTokenRepo{
		deleteFunc: func(refreshToken string) error {
			return expectedErr
		},
	}

	service := oauth2.NewOAuth2Service(conf, fakeRepo)

	err := service.Logout("valid_refresh_token")
	if err == nil || !errors.Is(err, expectedErr) {
		t.Errorf("ожидалась ошибка %v, получена %v", expectedErr, err)
	}
}
