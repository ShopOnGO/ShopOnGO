package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	_ "github.com/ShopOnGO/ShopOnGO/prod/docs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/req"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"

	googleOAuth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{
		Config:        deps.Config,
		AuthService:   deps.AuthService,
		OAuth2Service: deps.OAuth2Service,
	}
	router.HandleFunc("POST /auth/login", handler.Login())
	router.HandleFunc("GET /oauth/google/login", handler.GoogleLogin)
	router.HandleFunc("POST /auth/register", handler.Register())
	router.HandleFunc("POST /auth/logout", handler.Logout())
	router.Handle("POST /auth/change/password", middleware.IsAuthed(handler.ChangePassword(), deps.Config))
	router.Handle("POST /auth/change/role", middleware.IsAuthed(handler.ChangeUserRole(), deps.Config))
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

		role, err := h.AuthService.GetUserRole(email)
		if err != nil {
			http.Error(w, ErrFailedToGetUserRole+": "+err.Error(), http.StatusInternalServerError)
			return
		}

		jwtToken, refreshToken, err := h.OAuth2Service.GenerateTokens(email, role)
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
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {

	googleOauthConfig := &googleOAuth2.Config{
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
		url := googleOauthConfig.AuthCodeURL("state-token", googleOAuth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	// обменчик кода на токен
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, ErrFailedToExchangeToken+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, ErrFailedToGetUserInfo+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() //пока так

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, ErrFailedToDecodeUserInfo+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	role, err := h.AuthService.GetUserRole(userInfo.Email)
    if err != nil {
        http.Error(w, ErrFailedToGetUserInfo+": "+err.Error(), http.StatusInternalServerError)
        return
    }
	
	jwtToken, refreshToken, err := h.OAuth2Service.GenerateTokens(userInfo.Email, role)
	if err != nil {
		http.Error(w, ErrFailedToGenerateTokens+": "+err.Error(), http.StatusInternalServerError)
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

		role, err := h.AuthService.GetUserRole(email)
		if err != nil {
			http.Error(w, ErrFailedToGetUserRole+": "+err.Error(), http.StatusInternalServerError)
			return
		}

		jwtToken, refreshToken, err := h.OAuth2Service.GenerateTokens(email, role)
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
			http.Error(w, ErrFailedRefreshTokenNotFound, http.StatusUnauthorized)
			return
		}
		refreshToken := refreshCookie.Value

		// Вызываем метод logout сервиса, который удаляет refresh-токен
		if err := h.OAuth2Service.Logout(refreshToken); err != nil {
			http.Error(w, ErrFailedToLogout+": "+err.Error(), http.StatusInternalServerError)
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


// ChangePassword меняет пароль пользователя
// @Summary        Смена пароля
// @Description    Изменяет пароль пользователя, требует авторизации (Bearer токен)
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          body body ChangePasswordRequest true "Старый и новый пароль"
// @Success        200 {object} map[string]string "Сообщение об успешной смене пароля"
// @Failure        400 {string} string "Некорректные данные или старый пароль неверен"
// @Failure        401 {string} string "Неавторизован"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /auth/change/password [post]
func (h *AuthHandler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[ChangePasswordRequest](&w, r)
		if err != nil {
			return
		}

		// Извлекаем email пользователя из контекста (middleware.IsAuthed добавляет его)
		emailAny := r.Context().Value(middleware.ContextEmailKey)
		email, ok := emailAny.(string)
		if !ok || email == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := h.AuthService.ChangePassword(email, body.OldPassword, body.NewPassword); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res.Json(w, map[string]string{"message": "Password changed successfully"}, http.StatusOK)
	}
}

// ChangeUserRole изменяет роль пользователя
// @Summary        Изменение роли пользователя
// @Description    Изменяет роль пользователя, требует авторизации (Bearer токен)
// @Tags           auth
// @Accept         json
// @Produce        json
// @Param          body body ChangeRoleRequest true "Email пользователя и новая роль"
// @Success        200 {object} map[string]string "Сообщение об успешном изменении роли"
// @Failure        400 {string} string "Некорректные данные"
// @Failure        401 {string} string "Неавторизован"
// @Failure        403 {string} string "Недостаточно прав"
// @Failure        500 {string} string "Ошибка сервера"
// @Router         /auth/change/role [post]
func (h *AuthHandler) ChangeUserRole() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        body, err := req.HandleBody[ChangeRoleRequest](&w, r)
        if err != nil {
            http.Error(w, ErrInvalidRequestData, http.StatusBadRequest)
            return
        }

        // Обновляем роль пользователя в базе данных
        if err := h.AuthService.UpdateUserRole(body.Email, body.NewRole); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Генерируем новый JWT и refresh-токен с обновленной ролью
        jwtToken, refreshToken, err := h.OAuth2Service.GenerateTokens(body.Email, body.NewRole)
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

        res.Json(w, map[string]string{"message": "Role changed successfully", "token": jwtToken}, http.StatusOK)
    }
}