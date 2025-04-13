package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/jwt"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type contextKey string

const (
	ContextGuestIDKey contextKey = "guest_id"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

func AuthOrGuest(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Received request: %s %s", r.Method, r.URL.Path)
		authedHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authedHeader, "Bearer ") {
			token := strings.TrimPrefix(authedHeader, "Bearer ")
			isValid, data, err := jwt.NewJWT(config.OAuth.Secret).Parse(token)

			if err != nil || !isValid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				logger.Errorf("Invalid token: %v", err)
				return
			}
			logger.Infof("Token valid, UserID: %d, Role: %s", data.UserID, data.Role)

			// Токен валиден, добавляем userID в контекст
			ctx := context.WithValue(r.Context(), ContextUserIDKey, data.UserID)
			ctx = context.WithValue(ctx, ContextRolesKey, data.Role)

			session, _ := store.Get(r, "guest-session")
			if session.Values["guest_id"] != nil {
				if guestID, ok := session.Values["guest_id"].([]byte); ok {
					ctx = context.WithValue(ctx, ContextGuestIDKey, guestID)
					logger.Infof("Added guest_id from session: %v", guestID)
				} else {
					logger.Errorf("Failed to assert guest_id from session")
				}
			}
			
			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
			return
		}

		session, _ := store.Get(r, "guest-session")
		if session.Values["guest_id"] == nil {
			guestID := generateGuestID()
			session.Values["guest_id"] = guestID
			session.Save(r, w)
			logger.Infof("New guest session created with guest_id: %d", guestID)
		} else {
			// Логируем использование существующего guest_id
			logger.Infof("Found existing guest session with guest_id: %v", session.Values["guest_id"])
		}

		guestID, ok := session.Values["guest_id"].([]byte)
		if !ok {
			logger.Errorf("Failed to get valid guest_id from session")
			http.Error(w, "Ошибка получения guest_id", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), ContextGuestIDKey, guestID)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func generateGuestID() []byte {
    newUUID := uuid.New()
    return newUUID[:]
}