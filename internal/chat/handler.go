package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
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
		http.HandlerFunc(h.HandleWebSocket),
		deps.Config,
	)).Methods("GET")
	router.Handle("/api/chat/upload", middleware.IsAuthed(
		http.HandlerFunc(h.HandleFileUpload),
		deps.Config,
	)).Methods("POST")

}

func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	h.service.ServeWS(w, r)
}

func (h *ChatHandler) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем файл от клиента (ограничение 10МБ)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too big", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 2. Пересылаем файл в Media Service (он вернет нам JSON с URL)
	mediaServiceURL := "http://media_container:8084/media-service/uploads"

	uploadedURL, err := sendFileToMediaService(mediaServiceURL, file, header.Filename)
	if err != nil {
		// Логируем ошибку, чтобы видеть в консоли чата, что пошло не так
		logger.Errorf("Error sending to media service: %v\n", err)
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// 3. Определяем тип для фронтенда (картинка или файл)
	msgType := "file"
	ext := filepath.Ext(header.Filename)
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
		msgType = "image"
	}

	// 4. Отдаем ответ фронтенду чата
	resp := UploadResponse{
		URL:      uploadedURL, // Ссылка, которую вернул S3 через Media Service
		FileName: header.Filename,
		Type:     msgType,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Вспомогательная функция для проксирования файла
func sendFileToMediaService(targetURL string, file multipart.File, filename string) (string, error) {
	// Создаем буфер для тела запроса
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Создаем поле "file" внутри формы
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}

	// Копируем содержимое реального файла в форму
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	writer.Close()

	req, err := http.NewRequest("POST", targetURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("media service returned status: %d", resp.StatusCode)
	}

	// Читаем JSON ответ от Media Service: {"url": "..."}
	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}
