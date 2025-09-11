package chat

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/gorilla/mux"
)

type ChatHandlerDeps struct {
	ChatService *ChatService
	Config      *configs.Config
}

type ChatHandler struct {
	service *ChatService
}

func NewChatHandler(router *mux.Router, deps ChatHandlerDeps) {
	h := &ChatHandler{
		service: deps.ChatService,
	}
	router.Handle("/ws/chat", middleware.IsAuthed(
		http.HandlerFunc(h.HandleWebSocket), // Новое middleware
		deps.Config,
	)).Methods("GET")

}

func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	h.service.ServeWS(w, r)
}
