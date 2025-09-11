package chat

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все подключения (пока)
	},
}

type ChatService struct {
	hub  *Hub
	repo *ChatRepository
}

func NewChatService(repo *ChatRepository) *ChatService {
	hub := NewHub(repo)
	go hub.Run()
	return &ChatService{
		hub:  hub,
		repo: repo,
	}
}

func (c *ChatService) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Получаем userID из контекста
	userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверка роли
	role, ok := r.Context().Value(middleware.ContextRolesKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	fmt.Println(role)
	isManager := (role == "manager")
	fmt.Println(isManager)
	// Создаем клиента
	client := NewClient(conn, c.hub, userID, isManager)
	c.hub.register <- client
	// Запуск горутин для обработки сообщений
	go client.ReadPump()
	go client.WritePump()
}
