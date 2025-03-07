package auth

import (
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/jwt"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/req"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"
)

type AuthHandlerDeps struct { // содержит все необходимые элементы заполнения. это DC
	*configs.Config
	*AuthService
}
type AuthHandler struct { // это уже рабоая структура
	*configs.Config
	*AuthService
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{
		Config:      deps.Config,
		AuthService: deps.AuthService,
	}
	router.HandleFunc("POST /auth/login", handler.Login())
	router.HandleFunc("POST /auth/register", handler.Register())
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

		tokenManager := jwt.NewJWT(h.Config.Auth.Secret)
		
		jwtToken, err := tokenManager.Create(jwt.JWTData{Email: email}, time.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		refreshToken, err := tokenManager.NewRefreshToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправка refresh-токена в cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour), //типа месяц
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
		tokenManager := jwt.NewJWT(h.Config.Auth.Secret)

		jwtToken, err := tokenManager.Create(jwt.JWTData{Email: email}, time.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		refreshToken, err := tokenManager.NewRefreshToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})

		//fmt.Println(h.Config.Auth.Secret)
		data := LoginResponse{
			Token: jwtToken,
		}
		res.Json(w, data, 201)
	}
}
