package auth

import (
	"errors"
	"fmt"

	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"github.com/ShopOnGO/ShopOnGO/pkg/di"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
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
func (service *AuthService) Register(email, password, name string) (uint, error) {
	existedUser, err := service.UserRepository.FindByEmail(email)

	if err != nil && err != gorm.ErrRecordNotFound {
        return 0, err
    }

    if existedUser != nil {
        // Если пользователь найден, проверяем его провайдера
        if existedUser.Provider == "google" || existedUser.PasswordHash == "" {
            return 0, errors.New(ErrGoogleAuthToLocalFailed) // У Google-юзеров нет пароля
        }
        return 0, errors.New(ErrUserExists) // Пользователь с таким email уже существует
    }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //дефолтная cost даёт 2^10 раундов шифрования
	if err != nil {
		return 0, err
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
		return 0, err
	}
	return user.ID, nil
}

func (service *AuthService) Login(email, password string) (uint, error) {
	existedUser, _ := service.UserRepository.FindByEmail(email)
	if existedUser == nil {
		return 0, errors.New(ErrWrongCredentials)
	}

	if existedUser.Provider == "google" || existedUser.PasswordHash == "" {
		return 0, errors.New(ErrGoogleAuthToLocalFailed) // У Google-юзеров нет пароля
	}
	logger.Info("Сохраненный пароль в БД:", existedUser.PasswordHash)

	logger.Info("Введенный пароль: " + password)
	logger.Info("Хеш пароля из БД: " + existedUser.PasswordHash)

	err := bcrypt.CompareHashAndPassword([]byte(existedUser.PasswordHash), []byte(password)) //дефолтная cost даёт 2^10 раундов шифрования
	if err != nil {
		logger.Error("❌ Ошибка сравнения паролей: " + err.Error())
		return 0, errors.New(ErrWrongCredentials)
	}

	return existedUser.ID, nil
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
