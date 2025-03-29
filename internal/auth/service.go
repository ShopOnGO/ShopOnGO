package auth

import (
	"errors"
	"fmt"

	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/di"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"gorm.io/gorm"

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
	existedUser, err := service.UserRepository.FindByEmail(email)

	if err != nil && err != gorm.ErrRecordNotFound {
        return "", err
    }

    if existedUser != nil {
        // Если пользователь найден, проверяем его провайдера
        if existedUser.Provider == "google" || existedUser.PasswordHash == "" {
            return "", errors.New(ErrGoogleAuthToLocalFailed) // У Google-юзеров нет пароля
        }
        return "", errors.New(ErrUserExists) // Пользователь с таким email уже существует
    }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //дефолтная cost даёт 2^10 раундов шифрования
	if err != nil {
		return "", err
	}
	user := &user.User{
		Email:    email,
		PasswordHash: string(hashedPassword),
		Name:     name,
		Role: "buyer",
		Provider: "local",
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

	if existedUser.Provider == "google" || existedUser.PasswordHash == "" {
		return "", errors.New(ErrGoogleAuthToLocalFailed) // У Google-юзеров нет пароля
	}
	logger.Info("Сохраненный пароль в БД:", existedUser.PasswordHash)

	logger.Info("Введенный пароль: " + password)
	logger.Info("Хеш пароля из БД: " + existedUser.PasswordHash)

	err := bcrypt.CompareHashAndPassword([]byte(existedUser.PasswordHash), []byte(password)) //дефолтная cost даёт 2^10 раундов шифрования
	if err != nil {
		logger.Error("❌ Ошибка сравнения паролей: " + err.Error())
		return "", errors.New(ErrWrongCredentials)
	}

	return email, nil
}

func (service *AuthService) GetOrCreateUserByGoogle(userInfo GoogleUserInfo) (*user.User, error) {
	userInPostgres, err := service.UserRepository.FindByEmail(userInfo.Email)
	var role string
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если пользователь не найден, создаём нового
			role = "buyer"
			newUser := &user.User{
				Name:     userInfo.Name,
				Email:    userInfo.Email,
				Role:     role,
				Provider: "google",
			}
			createdUser, err := service.UserRepository.Create(newUser)
			if err != nil {
				return nil, err
			}
			return createdUser, nil
		}
		return nil, err
	}
	return userInPostgres, nil
}

func (service *AuthService) ChangePassword(email, oldPassword, newPassword string) error {
	userData, _ := service.UserRepository.FindByEmail(email)
	if userData == nil {
		return errors.New(ErrWrongCredentials)
	}

	if userData.Provider == "google" || userData.PasswordHash == "" {
		return errors.New(ErrWrongCredentials) // У Google-юзеров нет пароля
	}

	err := bcrypt.CompareHashAndPassword([]byte(userData.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New(ErrWrongPassword)
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

// Изменённый метод обновления данных пользователя
func (service *AuthService) UpdateUser(data *ChangeRoleRequest) error {
    userData, err := service.UserRepository.FindByEmail(data.Email)
    if err != nil {
        return fmt.Errorf(ErrFailedToFindUser+": %w", err)
    }
    if userData == nil {
        return errors.New(ErrUserNotFound)
    }

    userData.Role = data.NewRole
	userData.Phone = data.Phone
    if data.NewRole == "seller" {
        userData.StoreName = &data.StoreName
        userData.StoreAddress = &data.StoreAddress
        userData.StorePhone = &data.StorePhone
    }
    userData.AcceptTerms = data.AcceptTerms

	_, err = service.UserRepository.Update(userData)
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
