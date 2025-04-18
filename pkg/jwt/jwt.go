package jwt

import (
	"errors"
	"time"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

// TokenManager интерфейс для работы с JWT
type TokenManager interface {
	NewJWT(data JWTData, ttl time.Duration) (string, error)
	Parse(accessToken string) (bool, *JWTData, error)
}
type JWTData struct {
	UserID uint
	Role string
}
type JWT struct {
	Secret string
}

func NewJWT(secret string) *JWT {
	return &JWT{
		Secret: secret,
	}
}

func (j *JWT) Create(data JWTData, ttl time.Duration) (string, error) {
	if data.Role == "" {
		data.Role = "buyer"
	}

	//метод шифрования
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": data.UserID,
		"role": data.Role,
		"exp":   time.Now().Add(ttl).Unix(),
		//данные
	})
	s, err := t.SignedString([]byte(j.Secret)) // подпись
	if err != nil {
		return "", err
	}
	return s, nil
}

func (j *JWT) Parse(token string) (bool, *JWTData, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.Secret), nil // передача секрета для парсинга токена
	})
	if err != nil {
		logger.Error("�� Invalid token parse")
		return false, nil, err
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("�� Invalid token claims")
		return false, nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64) // JWT хранит числовые значения как float64
	if !ok {
		logger.Error("Invalid token: missing user_id")
		return false, nil, errors.New("invalid token: missing user_id")
	}
	userIDUint := uint(userID)

	role, ok := claims["role"].(string)
	if !ok {
		logger.Error("�� Invalid token: missing role")
		return false, nil, errors.New("invalid token: missing role")
	}

	return t.Valid, &JWTData{
		UserID: userIDUint,
		Role:  role,
	}, nil

}