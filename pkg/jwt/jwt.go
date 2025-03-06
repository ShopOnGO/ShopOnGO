package jwt

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager интерфейс для работы с JWT и Refresh токенами
type TokenManager interface {
	NewJWT(data JWTData, ttl time.Duration) (string, error)
	Parse(accessToken string) (bool, *JWTData, error)
	NewRefreshToken() (string, error)
}
type JWTData struct {
	Email string
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
	//метод шифрования
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": data.Email,
		"exp":   time.Now().Add(ttl).Unix(), // добавляем время жизни токена
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
		return false, nil, err
	}
	email, ok := t.Claims.(jwt.MapClaims)["email"]
	if !ok {
		return false, nil, errors.New("invalid token claims")
	}
	return t.Valid, &JWTData{
		Email: email.(string),
	}, nil

}

func (j *JWT) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	// создаем случайные байты для refresh токена
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}