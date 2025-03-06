package jwt_test

import (
	"testing"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/jwt"
)

func TestJWTCreate(t *testing.T) {
	const email = "a2@a.ru"
	jwtService := jwt.NewJWT("/2+XnmJGz1j3ehIVI/5P9kl+CghrE3DcS7rnT+qar5w=")
	token, err := jwtService.Create(jwt.JWTData{
		Email: email,
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	isValid, data, err := jwtService.Parse(token)
	if !isValid {
		t.Fatal("token invalid")
	}
	if err != nil {
		t.Fatal(err)
	}
	if data.Email != email {
		t.Fatalf("Email %s not equal %s", data.Email, email)
	}

}


func TestRefreshToken(t *testing.T) {
	jwtService := jwt.NewJWT("/2+XnmJGz1j3ehIVI/5P9kl+CghrE3DcS7rnT+qar5w=")

	refreshToken, err := jwtService.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем длину refresh токена (32 байта, кодированный в строку)
	if len(refreshToken) != 64 { // 32 байта = 64 символа в hex-формате
		t.Fatalf("expected refresh token of length 64, got %d", len(refreshToken))
	}
}