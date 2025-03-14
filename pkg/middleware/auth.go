package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/jwt"
)

type key string // делается чтобы не затирать другие значения в программе

const (
	ContextEmailKey key = "ContentEmailKey"
)

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}
func IsAuthed(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authedHeader, "Bearer") { // нужен ли пробел после Bearer?
			writeUnauthed(w)
			return
		}
		token := strings.TrimPrefix(authedHeader, "Bearer ")
		isValid, data, err := jwt.NewJWT(config.Auth.Secret).Parse(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !isValid {
			writeUnauthed(w)
			return
		}
		ctx := context.WithValue(r.Context(), ContextEmailKey, data.Email)
		req := r.WithContext(ctx) // для передачи контекста необходимо пересоздать запроc
		next.ServeHTTP(w, req)    //все handlers теперь обогащены необходимым контекстом
	})
}
