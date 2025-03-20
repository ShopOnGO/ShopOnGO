package auth

import (
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/req"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"
)

type AuthHandlerDeps struct {
	*configs.Config
	*AuthService
	OAuth2Service oauth2.OAuth2Service
}
type AuthHandler struct {
	*configs.Config
	*AuthService
	OAuth2Service oauth2.OAuth2Service
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{
		Config:        deps.Config,
		AuthService:   deps.AuthService,
		OAuth2Service: deps.OAuth2Service,
	}
	router.HandleFunc("POST /auth/login", handler.Login())
	router.HandleFunc("POST /auth/register", handler.Register())
	router.HandleFunc("POST /auth/logout", handler.Logout())
}


// Login аутентифицирует пользователя и выдает JWT токен
// @Summary        Вход в систему
// @Description    Аутентифицирует пользователя по email и паролю, возвращает JWT токен
// @Tags          auth
// @Accept        json
// @Produce       json
// @Param         body body LoginRequest true "Данные для входа"
// @Success       200 {object} LoginResponse "Успешный вход, возвращает JWT токен"
// @Failure       401 {string} string "Неверные учетные данные"
// @Failure       500 {string} string "Ошибка сервера при создании токена"
// @Router        /auth/login [post]
func (h *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[LoginRequest](&w, r)
		if err != nil {
			return
		}

		email, err := h.AuthService.Login(body.Email, body.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		jwtToken, refreshToken, err := h.OAuth2Service.GenerateTokens(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(h.Config.Redis.RefreshTokenTTL),
		})

		data := LoginResponse{
			Token: jwtToken,
		}
		res.Json(w, data, 200)
	}
}

// Register регистрирует нового пользователя и возвращает JWT токен
// @Summary        Регистрация нового пользователя
// @Description    Создает учетную запись пользователя и возвращает JWT токен для аутентификации
// @Tags          auth
// @Accept        json
// @Produce       json
// @Param         body body RegisterRequest true "Данные для регистрации"
// @Success       201 {object} LoginResponse "Успешная регистрация, возвращает JWT токен"
// @Failure       400 {string} string "Некорректные данные для регистрации"
// @Failure       409 {string} string "Пользователь с таким email уже существует"
// @Failure       500 {string} string "Ошибка сервера при создании токена"
// @Router        /auth/register [post]
func (h *AuthHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[RegisterRequest](&w, r)
		if err != nil {
			return
		}

		email, err := h.AuthService.Register(body.Email, body.Password, body.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		jwtToken, refreshToken, err := h.OAuth2Service.GenerateTokens(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(h.Config.Redis.RefreshTokenTTL),
		})

		//fmt.Println(h.Config.Auth.Secret)
		data := LoginResponse{
			Token: jwtToken,
		}
		res.Json(w, data, 201)
	}
}

// Logout завершает сеанс пользователя и удаляет refresh-токен из cookie
// @Summary        Завершение сеанса пользователя
// @Description    Удаляет refresh-токен из хранилища и очищает cookie
// @Tags          auth
// @Accept        json
// @Produce       json
// @Success       200 {object} map[string]string "Успешный выход, refresh-токен удален"
// @Failure       401 {string} string "Refresh-токен не найден"
// @Failure       500 {string} string "Ошибка сервера при выходе"
// @Router        /auth/logout [post]
func (h *AuthHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем refresh-токен из cookie
		refreshCookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, "Refresh token not found", http.StatusUnauthorized)
			return
		}
		refreshToken := refreshCookie.Value

		// Вызываем метод logout сервиса, который удаляет refresh-токен
		if err := h.OAuth2Service.Logout(refreshToken); err != nil {
			http.Error(w, "Failed to logout: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Очищаем cookie refresh-токена
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
		})

		res.Json(w, map[string]string{
			"message":      "Logout successful",
			"removeToken":  "Please remove access token from your storage",
		}, http.StatusOK)
	}
}
