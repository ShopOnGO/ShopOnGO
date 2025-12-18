package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/jwt"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
)

type key string

const (
	ContextUserIDKey key = "ContextUserIDKey"
	ContextRolesKey  key = "ContextRolesKey"
)

func ValidateToken(tokenString string, secret string) (uint, string, error) {
	isValid, data, err := jwt.NewJWT(secret).Parse(tokenString)

	if err != nil {
		return 0, "", err
	}
	if !isValid {
		return 0, "", errors.New("token is not valid")
	}

	return data.UserID, data.Role, nil
}

func IsAuthed(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authedHeader, "Bearer ") {
			logger.Error("❌ No valid Bearer prefix")
			writeUnauthed(w)
			return
		}

		token := strings.TrimPrefix(authedHeader, "Bearer ")

		userID, role, err := ValidateToken(token, config.OAuth.Secret)

		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				logger.Error("❌ Token expired:", err)
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			logger.Error("❌ Invalid token:", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		logger.Info("✅ Token is valid for:", userID)
		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
		logger.Info("Role:", role)
		ctx = context.WithValue(ctx, ContextRolesKey, role)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}
