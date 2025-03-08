package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/refresh"
)

type MockUserRepository struct{}

func (repo *MockUserRepository) Create(u *user.User) (*user.User, error) {
	return &user.User{
		Email: "a@a.ru",
	}, nil

}
func (repo *MockUserRepository) FindByEmail(email string) (*user.User, error) {
	return nil, nil
}


type MockAuthRepository struct{}

func (repo *MockAuthRepository) GetRefreshToken(token string) (*refresh.RefreshTokenRecord, error) {
	if token == "valid_token" {
		return &refresh.RefreshTokenRecord{
			Token:     token,
			Email:     "a@a.ru",
			ExpiresAt: time.Now().Add(time.Hour), // Токен ещё действителен
			IsRevoked: false,
		}, nil
	}
	return nil, errors.New("token not found")
}

func (repo *MockAuthRepository) RevokeRefreshToken(token string) error {
	if token == "valid_token" {
		return nil
	}
	return errors.New("failed to revoke")
}

func (repo *MockAuthRepository) SaveRefreshToken(token string, email string, expiresAt time.Time) error {
	return nil
}

// Тест на регистрацию пользователя
func TestRegisterSuccess(t *testing.T) {
	authService := auth.NewAuthService(&MockUserRepository{}, &MockAuthRepository{})

	const initialEmail = "a@a.ru"
	email, err := authService.Register(initialEmail, "password123", "dan")
	if err != nil {
		t.Fatal(err)
	}

	if email != initialEmail {
		t.Fatalf("email %s do not match expected %s", email, initialEmail)
	}
	// по сути обычный юнит тест для регистрации, только делаем вид записи в базу
	// для написания просто изучаем функцию и имитируем работу ее зависимостей
}

// можно добавить тесты на:
// Тест на успешное обновление refresh-токена
// Тест на просроченный токен
// Тест на отзыв токена