package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/jwt"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
)

type key string // делается чтобы не затирать другие значения в программе

const (
	ContextUserIDKey key = "ContextUserIDKey"
	ContextRolesKey key = "ContextRolesKey"
)

func IsAuthed(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authedHeader, "Bearer ") {
			logger.Error("❌ No valid Bearer prefix")
			writeUnauthed(w)
			return
		}
		token := strings.TrimPrefix(authedHeader, "Bearer ")
		isValid, data, err := jwt.NewJWT(config.OAuth.Secret).Parse(token)
		logger.Error("Received token:", token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				logger.Error("❌ Token expired:", err)
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			logger.Error("❌ Invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !isValid {
			logger.Error("❌ Token is not valid")
			writeUnauthed(w)
			return
		}
		logger.Info("✅ Token is valid for:", data.UserID)
		ctx := context.WithValue(r.Context(), ContextUserIDKey, data.UserID)
		logger.Info("Role:", data.Role)
		ctx = context.WithValue(ctx, ContextRolesKey, data.Role)
		req := r.WithContext(ctx) // для передачи контекста необходимо пересоздать запроc
		next.ServeHTTP(w, req)    //все handlers теперь обогащены необходимым контекстом
	})
}

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}