package oauth2

import (
	"encoding/json"
	"net/http"
	"time"

)

// OAuth2Handler отвечает за HTTP‑эндпоинты OAuth2.
type OAuth2HandlerDeps struct {
	Service OAuth2Service
}

type OAuth2Handler struct {
	service OAuth2Service
}

func NewOAuth2Handler(router *http.ServeMux, deps OAuth2HandlerDeps) {
	handler := &OAuth2Handler{
		service: deps.Service,
	}

	router.HandleFunc("/oauth/token", handler.HandleToken)
	router.HandleFunc("/oauth/authorize", handler.HandleAuthorize)
}


// HandleToken обрабатывает запрос обновления токенов.
func (h *OAuth2Handler) HandleToken(w http.ResponseWriter, r *http.Request) {
	// Читаем refresh‑токен из cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token cookie not found", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value

	accessToken, newRefreshToken, err := h.service.RefreshTokens(refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Обновляем cookie с новым refresh‑токеном
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})

	response := map[string]string{"access_token": accessToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleAuthorize обрабатывает запросы авторизации.
func (h *OAuth2Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OAuth2 Authorization Page"))
}
