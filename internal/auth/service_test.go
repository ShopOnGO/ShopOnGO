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

func (repo *MockUserRepository) Update(*user.User) (*user.User, error) {
	return nil, nil
}

func (repo *MockUserRepository) Delete(id uint) error {
	return nil
}

func (repo *MockUserRepository) UpdateUserPassword(id uint, newPassword string) error {
	return nil
}

func (repo *MockUserRepository) GetUserRoleByEmail(email string) (string, error) {
    return "", nil
}

func (repo *MockUserRepository) UpdateRole(user *user.User, role string) error {
	return nil
}

// Тест на регистрацию пользователя
func TestRegisterSuccess(t *testing.T) {
	authService := auth.NewAuthService(&MockUserRepository{})

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
