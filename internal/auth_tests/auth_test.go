package auth_test

import (
	"errors"
	"testing"

	"github.com/ShopOnGO/ShopOnGO/internal/auth"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockUserRepository реализует интерфейс di.IUserRepository
type MockUserRepository struct {
	// Мы будем задавать эти функции в каждом тесте,
	// чтобы контролировать, что "база данных" возвращает.
	findByEmailFunc        func(email string) (*user.User, error)
	createFunc             func(u *user.User) (*user.User, error)
	updateFunc             func(u *user.User) (*user.User, error)
	getUserRoleByEmailFunc func(email string) (string, error)
}

func (m *MockUserRepository) Create(u *user.User) (*user.User, error) {
	if m.createFunc != nil {
		return m.createFunc(u)
	}
	// Поведение по умолчанию, если функция не задана
	u.ID = 1
	return u, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*user.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(u *user.User) (*user.User, error) {
	if m.updateFunc != nil {
		return m.updateFunc(u)
	}
	return u, nil
}

func (m *MockUserRepository) GetUserRoleByEmail(email string) (string, error) {
	if m.getUserRoleByEmailFunc != nil {
		return m.getUserRoleByEmailFunc(email)
	}
	return "buyer", nil
}

func (m *MockUserRepository) Delete(id uint) error {
	return nil
}
func (m *MockUserRepository) UpdateUserPassword(id uint, newPassword string) error {
	return nil
}
func (m *MockUserRepository) UpdateRole(user *user.User, role string) error {
	return nil
}

func hashPassword(t *testing.T, password string) string {
	// require прервет тест, если хеширование не удастся
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password for test setup")
	return string(hash)
}

//  Тест AuthService.Register

func TestAuthService_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createFunc: func(u *user.User) (*user.User, error) {
				u.ID = 123
				return u, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		userID, err := service.Register("new@user.com", "password123", "New User")

		assert.NoError(t, err)
		assert.Equal(t, uint(123), userID)
	})

	t.Run("Failure - User Exists (local)", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return &user.User{
					Model: gorm.Model{
						ID: 1,
					},
					Email:        "exists@user.com",
					PasswordHash: "somehash",
					Provider:     "local",
				}, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		userID, err := service.Register("exists@user.com", "password123", "Test User")

		assert.Error(t, err) // Ошибка должна быть
		assert.Equal(t, auth.ErrUserExists, err.Error())
		assert.Equal(t, uint(0), userID) // ID должен быть 0
	})

	t.Run("Failure - User Exists (google)", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				// Пользователь найден, и он "google"
				return &user.User{
					Model: gorm.Model{
						ID: 1,
					},
					Email:        "google@user.com",
					Provider:     "google",
					PasswordHash: "", // У Google-юзера нет хеша пароля
				}, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		userID, err := service.Register("google@user.com", "password123", "Test User")

		assert.Error(t, err)
		assert.Equal(t, auth.ErrGoogleAuthToLocalFailed, err.Error())
		assert.Equal(t, uint(0), userID)
	})

	t.Run("Failure - DB error on FindByEmail", func(t *testing.T) {
		dbError := errors.New("connection lost")
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				// Имитируем ошибку БД, которая не "запись не найдена"
				return nil, dbError
			},
		}
		service := auth.NewAuthService(mockRepo)

		_, err := service.Register("new@user.com", "password123", "Test User")

		assert.Error(t, err)
		assert.Equal(t, dbError, err) // Ошибка должна быть проброшена "как есть"
	})

	t.Run("Failure - DB error on Create", func(t *testing.T) {
		dbError := errors.New("failed to create")
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createFunc: func(u *user.User) (*user.User, error) {
				return nil, dbError
			},
		}
		service := auth.NewAuthService(mockRepo)

		_, err := service.Register("new@user.com", "password123", "Test User")

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
	})
}

//  Тест AuthService.Login

func TestAuthService_Login(t *testing.T) {
	const correctPassword = "password123"
	hashedPassword := hashPassword(t, correctPassword)

	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return &user.User{
					Model: gorm.Model{
						ID: 42,
					},
					Email:        "login@user.com",
					PasswordHash: hashedPassword,
					Provider:     "local",
				}, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		userID, err := service.Login("login@user.com", correctPassword)

		assert.NoError(t, err)
		assert.Equal(t, uint(42), userID)
	})

	t.Run("Failure - User Not Found", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		_, err := service.Login("notfound@user.com", "password123")

		assert.Error(t, err)
		assert.Equal(t, auth.ErrWrongCredentials, err.Error())
	})

	t.Run("Failure - Wrong Password", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return &user.User{
					Model: gorm.Model{
						ID: 42,
					},
					Email:        "login@user.com",
					PasswordHash: hashedPassword,
					Provider:     "local",
				}, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		// Попытка войти с "wrongpassword"
		_, err := service.Login("login@user.com", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, auth.ErrWrongCredentials, err.Error())
	})

	t.Run("Failure - Google User", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return &user.User{
					Model: gorm.Model{
						ID: 42,
					},
					Email:        "google@user.com",
					PasswordHash: "",
					Provider:     "google",
				}, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		_, err := service.Login("google@user.com", "password123")

		assert.Error(t, err)
		assert.Equal(t, auth.ErrGoogleAuthToLocalFailed, err.Error())
	})
}

//  Тест AuthService.GetOrCreateUserByGoogle

func TestAuthService_GetOrCreateUserByGoogle(t *testing.T) {
	googleInfo := auth.GoogleUserInfo{
		ID:    "google123",
		Name:  "Google User",
		Email: "google@user.com",
	}

	t.Run("Success - Get Existing User", func(t *testing.T) {
		existingUser := &user.User{
			Model: gorm.Model{
				ID: 77,
			},
			Email:    "google@user.com",
			Name:     "Google User",
			Provider: "google",
			Role:     "seller",
		}
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				assert.Equal(t, googleInfo.Email, email)
				return existingUser, nil
			},
			// Create не должен быть вызван
			createFunc: func(u *user.User) (*user.User, error) {
				t.Fatal("Create should not be called")
				return nil, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		u, err := service.GetOrCreateUserByGoogle(googleInfo)

		assert.NoError(t, err)
		assert.Equal(t, existingUser.ID, u.ID)
		assert.Equal(t, "seller", u.Role) // Роль осталась "seller"
	})

	t.Run("Success - Create New User", func(t *testing.T) {
		newUser := &user.User{
			Model: gorm.Model{
				ID: 1,
			},
			Email:    googleInfo.Email,
			Name:     googleInfo.Name,
			Provider: "google",
			Role:     "buyer",
		}
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createFunc: func(u *user.User) (*user.User, error) {
				// Проверяем, что создается правильный пользователь
				assert.Equal(t, googleInfo.Email, u.Email)
				assert.Equal(t, googleInfo.Name, u.Name)
				assert.Equal(t, "google", u.Provider)
				assert.Equal(t, "buyer", u.Role)
				u.ID = 1
				return u, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		u, err := service.GetOrCreateUserByGoogle(googleInfo)

		assert.NoError(t, err)
		assert.Equal(t, newUser.ID, u.ID)
		assert.Equal(t, "buyer", u.Role)
	})

	t.Run("Failure - DB error on Find", func(t *testing.T) {
		dbError := errors.New("db find error")
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, dbError
			},
		}
		service := auth.NewAuthService(mockRepo)

		_, err := service.GetOrCreateUserByGoogle(googleInfo)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
	})
}

//  Тест AuthService.UpdateUser

func TestAuthService_UpdateUser(t *testing.T) {
	changeRequest := &auth.ChangeRoleRequest{
		Email:        "user@example.com",
		NewRole:      "seller",
		Phone:        "123456789",
		StoreName:    "My Store",
		StoreAddress: "123 Main St",
		StorePhone:   "987654321",
		AcceptTerms:  true,
	}

	existingUser := &user.User{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "user@example.com",
		Role:  "buyer",
	}

	t.Run("Success - Update to Seller", func(t *testing.T) {
		var updatedUser *user.User
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return existingUser, nil
			},
			updateFunc: func(u *user.User) (*user.User, error) {
				updatedUser = u
				return u, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		err := service.UpdateUser(changeRequest)

		assert.NoError(t, err)
		require.NotNil(t, updatedUser)
		assert.Equal(t, "seller", updatedUser.Role)
		assert.Equal(t, "123456789", updatedUser.Phone)
		assert.Equal(t, "My Store", *updatedUser.StoreName)
		assert.Equal(t, "123 Main St", *updatedUser.StoreAddress)
		assert.Equal(t, "987654321", *updatedUser.StorePhone)
		assert.True(t, updatedUser.AcceptTerms)
	})

	t.Run("Success - Update to Buyer (Store fields not set)", func(t *testing.T) {
		changeRequestBuyer := &auth.ChangeRoleRequest{
			Email:   "user@example.com",
			NewRole: "buyer", // Меняем на 'buyer'
			Phone:   "11111",
		}
		// У этого юзера уже есть 'StoreName'
		existingSeller := &user.User{
			Model: gorm.Model{
				ID: 1,
			},
			Email:     "user@example.com",
			Role:      "seller",
			StoreName: new(string),
		}
		*existingSeller.StoreName = "Old Store"

		var updatedUser *user.User
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return existingSeller, nil
			},
			updateFunc: func(u *user.User) (*user.User, error) {
				updatedUser = u
				return u, nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		err := service.UpdateUser(changeRequestBuyer)

		assert.NoError(t, err)
		assert.Equal(t, "buyer", updatedUser.Role)  // Роль обновилась
		assert.Equal(t, "11111", updatedUser.Phone) // Телефон обновился
		assert.Equal(t, "Old Store", *updatedUser.StoreName)
	})

	t.Run("Failure - User Not Found", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, nil
			},
		}
		service := auth.NewAuthService(mockRepo)
		err := service.UpdateUser(changeRequest)

		assert.Error(t, err)
		assert.Equal(t, auth.ErrUserNotFound, err.Error())
	})

	t.Run("Failure - DB error on Find", func(t *testing.T) {
		dbError := errors.New("db find error")
		mockRepo := &MockUserRepository{
			findByEmailFunc: func(email string) (*user.User, error) {
				return nil, dbError
			},
		}
		service := auth.NewAuthService(mockRepo)
		err := service.UpdateUser(changeRequest)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), auth.ErrFailedToFindUser)
		assert.ErrorIs(t, err, dbError)
	})
}

//  Тест для AuthService.GetUserRole

func TestAuthService_GetUserRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			getUserRoleByEmailFunc: func(email string) (string, error) {
				return "admin", nil
			},
		}
		service := auth.NewAuthService(mockRepo)

		role, err := service.GetUserRole("admin@user.com")

		assert.NoError(t, err)
		assert.Equal(t, "admin", role)
	})

	t.Run("Failure - DB Error", func(t *testing.T) {
		dbError := errors.New("role not found")
		mockRepo := &MockUserRepository{
			getUserRoleByEmailFunc: func(email string) (string, error) {
				return "", dbError
			},
		}
		service := auth.NewAuthService(mockRepo)

		role, err := service.GetUserRole("admin@user.com")

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Equal(t, "", role)
	})
}
