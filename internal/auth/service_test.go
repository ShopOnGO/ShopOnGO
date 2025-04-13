package auth_test

import (
	"testing"

	"github.com/ShopOnGO/ShopOnGO/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
)

type MockUserRepository struct{}

func (repo *MockUserRepository) Create(u *user.User) (*user.User, error) {
    u.ID = 1
    return u, nil
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

	const expectedID uint = 1
	userID, err := authService.Register("a@a.ru", "password123", "dan")
	if err != nil {
		t.Fatal(err)
	}

	if userID != expectedID {
		t.Fatalf("userID %d does not match expected %d", userID, expectedID)
	}
}
