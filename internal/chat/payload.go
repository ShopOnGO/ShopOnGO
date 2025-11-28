package chat

// Команда от менеджера: взять пользователя, закрыть сессию и т.п.
type ManagerCommand struct {
	Command string `json:"command"`
	UserID  uint   `json:"user_id,omitempty"`
}

// Обычное текстовое сообщение от менеджера пользователю
type ManagerMessage struct {
	UserID   uint   `json:"user_id"`
	Content  string `json:"content"`
	Type     string `json:"type"`      // <-- Добавили
	FileName string `json:"file_name"` // <-- Добавили
}

// Общий ответ сервера
type ServerResponse struct {
	Status  string      `json:"status"` // "success" или "error"
	Message string      `json:"message,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// UploadResponse ответ с URL файла
type UploadResponse struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	Type     string `json:"type"`
}
type IncomingUserMessage struct {
	Content  string `json:"content"`
	Type     string `json:"type"`      // "text", "image", "file"
	FileName string `json:"file_name"` // Опционально
}
