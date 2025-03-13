package oauth2server

import (
	"fmt"
	"net/http"

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

// HandleToken выдаёт Access и Refresh токены по grant_type (password, refresh_token и т. д.).
func (s *OAuth2Server) HandleToken(w http.ResponseWriter, r *http.Request) {
	err := s.server.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func (s *OAuth2Server) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/oauth/token", s.HandleToken)
	mux.HandleFunc("/oauth/authorize", s.HandleAuthorize)
	mux.HandleFunc("/oauth/authpage", s.HandleAuthPage) // Если понадобится страница авторизации
}
