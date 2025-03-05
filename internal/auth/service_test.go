package auth_test

import (
	"testing"

	"github.com/ShopOnGO/ShopOnGO/prod/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
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

func TestRegisterSuccess(t *testing.T) {
	const initialEmail = "a@a.ru"
	authService := auth.NewAuthService(&MockUserRepository{})
	email, err := authService.Register(initialEmail, "1", "dan")
	if err != nil {
		t.Fatal(err)
	}
	if email != initialEmail {
		t.Fatalf("email %s do not match %s", email, initialEmail)
	}
	// по сути обычный юнит тест для регистрации, только делаем вид записи в базу
	// для написания просто изучаем функцию и имитируем работу ее зависимостей
}
