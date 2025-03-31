package middleware

import (
	"context"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/google/uuid"
)

const ContextGuestIDKey key = "ContextGuestIDKey"

func IsGuest(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, есть ли уже userID в контексте (значит, пользователь авторизован)
		if userID, ok := r.Context().Value(ContextUserIDKey).(uint); ok && userID != 0 {
			next.ServeHTTP(w, r) // Если авторизован, пропускаем дальше
			return
		}

		cookie, err := r.Cookie("guestID")
		var guestID string
		if err != nil {
			guestID = GenerateGuestID()
			http.SetCookie(w, &http.Cookie{
				Name:     "guestID",
				Value:    guestID,
				HttpOnly: true,
				Path:     "/",
			})
		} else {
			guestID = cookie.Value
			if _, err := uuid.Parse(guestID); err != nil {
				http.Error(w, "Invalid guestID", http.StatusBadRequest)
				return
			}
		}

		ctx := context.WithValue(r.Context(), ContextGuestIDKey, guestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsAuthedOrGuest(next http.Handler, config *configs.Config) http.Handler {
	return IsAuthed(IsGuest(next, config), config)
}


func GenerateGuestID() string {
	return uuid.NewString() // Генерируем строковый UUID
}