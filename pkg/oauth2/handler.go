package oauth2

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)
type OAuth2HandlerDeps struct {
	Service OAuth2Service
	*configs.Config
}

type OAuth2Handler struct {
	service OAuth2Service
	*configs.Config
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

func NewOAuth2Handler(router *http.ServeMux, deps OAuth2HandlerDeps) {
	handler := &OAuth2Handler{
		service: deps.Service,
		Config:  deps.Config,
	}

	router.HandleFunc("/oauth/token", handler.HandleToken)
	router.HandleFunc("/oauth/google/login", handler.GoogleLogin)
}


// HandleToken обновляет access-токен по refresh-токену
// @Summary        Обновление access-токена
// @Description    Обновляет access-токен, используя refresh-токен из cookie
// @Tags           auth
// @Accept         json
// @Produce        json
// @Success        200 {object} map[string]string "Новый access-токен"
// @Failure        401 {string} string "Refresh-токен отсутствует или недействителен"
// @Failure        500 {string} string "Ошибка сервера при обновлении токена"
// @Router         /oauth/token [post]
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
		Expires:  time.Now().Add(h.Redis.RefreshTokenTTL),
	})

	response := map[string]string{"access_token": accessToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


// GoogleLogin выполняет аутентификацию пользователя через Google OAuth2
// @Summary        Авторизация через Google
// @Description    Перенаправляет пользователя на страницу авторизации Google, затем получает токены и информацию о пользователе
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          code query string false "Код авторизации от Google (автоматически передается после редиректа)"
// @Success        200 {object} map[string]string "JWT access-токен"
// @Failure        500 {string} string "Ошибка при обмене кода на токен или получении данных пользователя"
// @Router         /oauth/google/login [get]
func (h *OAuth2Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {

	googleOauthConfig := &oauth2.Config{
		ClientID:     h.Config.Google.ClientID,
		ClientSecret: h.Config.Google.ClientSecret,
		RedirectURL:  h.Config.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Если параметр code отсутствует, перенаправляем пользователя на страницу согласия Google
	code := r.URL.Query().Get("code")
	if code == "" {
		// можно добавить state для безопасности, пока используется простая строка "state-token"
		url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	// обменчик кода на токен
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() //пока так

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jwtToken, refreshToken, err := h.service.GenerateTokens(userInfo.Email)
	if err != nil {
		http.Error(w, "Failed to generate tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(h.Config.Redis.RefreshTokenTTL),
	})

	response := map[string]string{
		"access_token": jwtToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
