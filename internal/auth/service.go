package auth

import (
	"errors"

	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/di"

	"golang.org/x/crypto/bcrypt"
)


type AuthService struct {
	UserRepository di.IUserRepository // измненено с *user.UserRepository для тестирования
}

func NewAuthService(userRepository di.IUserRepository) *AuthService {
	return &AuthService{
		UserRepository: userRepository,
	}
}

// Methods
func (service *AuthService) Register(email, password, name string) (string, error) {
	existedUser, _ := service.UserRepository.FindByEmail(email)
	if existedUser != nil {
		return "", errors.New(ErrUserExists)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //дефолтная cost даёт 2^10 раундов шифрования
	if err != nil {
		return "", err
	}
	user := &user.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}

	_, err = service.UserRepository.Create(user)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}

func (service *AuthService) Login(email, password string) (string, error) {
	existedUser, _ := service.UserRepository.FindByEmail(email)
	if existedUser == nil {

		return "", errors.New(ErrWrongCredentials)
	}
	err := bcrypt.CompareHashAndPassword([]byte(existedUser.Password), []byte(password)) //дефолтная cost даёт 2^10 раундов шифрования
	if err != nil {
		return "", errors.New(ErrWrongCredentials)
	}
	return email, nil
}

//пока что просто заглушка, надо решить как реализовать
func (service *AuthService) Refresh(refreshToken string) (string, error) {
	return "12345", nil
}
