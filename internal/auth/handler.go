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

// Допустим, refreshInput используется, если вы хотите принимать refresh-токен из JSON.
// Если же вы берёте его из cookie, то структура не обязательна.
type refreshInput struct {
	Token string `json:"token"`
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{
		Config:      deps.Config,
		AuthService: deps.AuthService,
	}
	router.HandleFunc("POST /auth/login", handler.Login())
	router.HandleFunc("POST /auth/register", handler.Register())

	router.HandleFunc("POST /auth/refresh", handler.Refresh())
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



// Refresh обновляет JWT токен, используя refresh-токен
// @Summary        Обновление токенов
// @Description    Принимает refresh-токен (например, из cookie), проверяет его и возвращает новый JWT токен
// @Tags           auth
// @Accept         json
// @Produce        json
// @Success        200 {object} LoginResponse "Новый JWT токен"
// @Failure        400 {string} string "Некорректный запрос"
// @Failure        401 {string} string "Неверный или просроченный refresh-токен"
// @Failure        500 {string} string "Ошибка сервера при создании токена"
// @Router         /auth/refresh [post]
func (h *AuthHandler) Refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем refresh-токен из cookie как пример
		cookie, err := r.Cookie("refresh_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Refresh token not found", http.StatusUnauthorized)
			return
		}
		refreshToken := cookie.Value

		// Проверка refresh-токена.
		// Здесь необходимо реализовать логику валидации refresh-токена.
		// Например, можно добавить метод h.AuthService.Refresh(refreshToken) который:
		// - Проверяет, существует ли такой токен в БД
		// - Возвращает email или идентификатор пользователя, если токен действителен
		// - Если токен недействителен, обработчик возвращает 401.
		email, err := h.AuthService.Refresh(refreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Создаем новый access-токен с временем жизни 1 час, надо бы константу завести что-ли...
		tokenManager := jwt.NewJWT(h.Config.Auth.Secret)
		jwtToken, err := tokenManager.Create(jwt.JWTData{Email: email}, time.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Генерируем новый refresh-токен
		newRefreshToken, err := tokenManager.NewRefreshToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Обновляем refresh-токен в cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})

		// Возвращаем новый JWT токен клиенту
		data := LoginResponse{
			Token: jwtToken,
		}
		res.Json(w, data, http.StatusOK)
	}
}