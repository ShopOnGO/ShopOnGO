package auth

import (
	"errors"
	"fmt"

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

func (service *AuthService) ChangePassword(email, oldPassword, newPassword string) error {
	userData, _ := service.UserRepository.FindByEmail(email)
	if userData == nil {
		return errors.New(ErrWrongCredentials)
	}

	err := bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(oldPassword))
	if err != nil {
		return errors.New(ErrWrongCredentials)
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

	if err != nil {
		return fmt.Errorf(FailedToHashNewPassword+": %w", err)
	}

	if err := service.UserRepository.UpdateUserPassword(userData.ID, string(newPasswordHash)); err != nil {
		return fmt.Errorf(FailedToUpdatePassword+": %w", err)
	}

	return nil
}

func (service *AuthService) UpdateUserRole(email, newRole string) error {
    userData, err := service.UserRepository.FindByEmail(email)
    if err != nil {
        return fmt.Errorf(ErrFailedToFindUser+": %w", err)
    }

    if userData == nil {
        return errors.New(ErrUserNotFound)
    }

    err = service.UserRepository.UpdateRole(userData, newRole)
    if err != nil {
        return fmt.Errorf(ErrFailedToUpdateUserRole+": %w", err)
    }

    return nil
}

func (service *AuthService) GetUserRole(email string) (string, error) {
	role, err := service.UserRepository.GetUserRoleByEmail(email)
    if err != nil {
        return "", err
    }
    return role, nil
}