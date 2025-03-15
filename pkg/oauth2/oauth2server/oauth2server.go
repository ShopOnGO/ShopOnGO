package oauth2server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2/oauth2manager"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"
)

// OAuth2Server — структура для запуска сервера OAuth2.
type OAuth2Server struct {
	oauthManager *oauth2manager.OAuth2ManagerImpl
	server       *server.Server
}

// NewOAuth2Server создает и настраивает сервер OAuth2.
func NewOAuth2Server(manager *oauth2manager.OAuth2ManagerImpl) *OAuth2Server {
	srv := server.NewServer(server.NewConfig(), manager.Manager)

	// Настроим обработку ошибок
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		fmt.Println("OAuth2 internal error:", err)
		return nil
	})
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		fmt.Println("OAuth2 response error:", re.Error.Error())
	})

	return &OAuth2Server{
		oauthManager: manager,
		server:       srv,
	}
}


func (s *OAuth2Server) HandleToken(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token cookie not found", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value

	// Проверяем, есть ли такой refresh_token в Redis
	userID, err := s.oauthManager.GetUserIDByRefreshToken(refreshToken)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Генерируем новые токены
	accessToken, newRefreshToken, err := s.oauthManager.GenerateTokens(userID)

	if err != nil {
		http.Error(w, "Failed to generate new tokens", http.StatusInternalServerError)
		return
	}

	// Обновляем refresh_token в Redis
	err = s.oauthManager.StoreRefreshToken(userID, newRefreshToken, 30*24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем новый refresh_token в куку
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})

	// Возвращаем новый access_token клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"access_token": "%s"}`, accessToken)))
}




// HandleAuthPage — эндпоинт для страницы авторизации (если нужен).
func (s *OAuth2Server) HandleAuthPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OAuth2 Authorization Page")) // Тут можно рендерить HTML-страницу.
}

// HandleAuthorize — обрабатывает запросы на выдачу кода авторизации (если используется Authorization Code Flow).
func (s *OAuth2Server) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	err := s.server.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}