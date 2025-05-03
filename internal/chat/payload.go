package chat

// Команда от менеджера: взять пользователя, закрыть сессию и т.п.
type ManagerCommand struct {
	Command string `json:"command"`
	UserID  uint   `json:"user_id,omitempty"`
}

// Обычное текстовое сообщение от менеджера пользователю
type ManagerMessage struct {
	UserID  uint   `json:"user_id"`
	Content string `json:"content"`
}

// Общий ответ сервера
type ServerResponse struct {
	Status  string      `json:"status"` // "success" или "error"
	Message string      `json:"message,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}
