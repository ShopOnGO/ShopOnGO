package refresh

import (
	"time"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
)

type AuthRepository struct {
	Database *db.Db
}

func NewAuthRepository(database *db.Db) *AuthRepository {
	return &AuthRepository{
		Database: database,
	}
}

// GetRefreshToken ищет refresh-токен в БД
func (repo *AuthRepository) GetRefreshToken(token string) (*RefreshTokenRecord, error) {
	var record RefreshTokenRecord
	result := repo.Database.DB.First(&record, "token = ?", token)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// SaveRefreshToken сохраняет новый refresh-токен в БД
func (repo *AuthRepository) SaveRefreshToken(token, email string, expiresAt time.Time) error {
	refreshToken := RefreshTokenRecord{
		Token:     token,
		Email:     email,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
	result := repo.Database.DB.Create(&refreshToken)
	return result.Error
}

// RevokeRefreshToken помечает токен как отозванный (is_revoked = true)
func (repo *AuthRepository) RevokeRefreshToken(token string) error {
	result := repo.Database.DB.Model(&RefreshTokenRecord{}).
		Where("token = ?", token).
		Update("is_revoked", true)
	return result.Error
}
