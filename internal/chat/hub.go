package chat

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
)

var sessionIDCounter uint64

type Session struct {
	ID          uint // Уникальный ID сессии\
	UserID      uint
	UserClients map[*Client]bool
	Manager     *Client
	History     []*Message
	mu          sync.RWMutex
}

type Hub struct {
	clients         map[*Client]bool    // no mutex
	sessions        map[uint]*Session   // mutex // по id пользователя
	managerSessions map[uint][]*Session // по id managera
	waitingUsers    map[uint]bool       // вместо map[uint]*Session
	register        chan *Client        // mutex
	unregister      chan *Client        // mutex
	ChatRepository  *ChatRepository     // Репозиторий для работы с сообщениями
	mu              sync.RWMutex
} // идеал был переписать на sync.map чтобы не забывать мьютексы

func NewHub(chatRepository *ChatRepository) *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		sessions:        make(map[uint]*Session), // id пользователей
		managerSessions: make(map[uint][]*Session),
		waitingUsers:    make(map[uint]bool),
		register:        make(chan *Client, 100),
		unregister:      make(chan *Client, 100), // каналы потокобезопасны!!!
		ChatRepository:  chatRepository,
	}
}

func (h *Hub) Run() {

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.handleNewClient(client)

		case client := <-h.unregister:
			h.removeClient(client)
		}
	}
}

func (h *Hub) handleNewClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if c.IsManager {
		// Менеджер сам выберет кого взять
		return
	}

	// Ищем сессию по ID пользователя
	session, exists := h.sessions[c.ID]
	if !exists {
		// Сессии нет - создаем новую
		session = &Session{
			ID:          generateUniqueID(),
			UserID:      c.ID,
			UserClients: make(map[*Client]bool),
			History:     []*Message{},
		}
		session.UserClients[c] = true
		h.sessions[c.ID] = session

		// Загружаем историю сообщений для новой сессии
		err := h.LoadLastMessages(c.ID, session, 50)
		if err != nil {
			logger.Error("LoadLastMessages failed", err)
		}

	} else {
		// Сессия уже существует - просто добавляем нового клиента в пул
		session.mu.Lock()
		session.UserClients[c] = true
		session.mu.Unlock()
	}
	sendHistoryToClient(c, session.History)
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c]; !ok {
		return // Клиент уже удалён
	}
	delete(h.clients, c) // Удаляем клиента из общего списка

	if c.IsManager {
		// Обработка отключения менеджера
		if sessions, ok := h.managerSessions[c.ID]; ok {
			for _, session := range sessions {
				if session.UserID != 0 {
					// Возвращаем пользователя в очередь ожидания
					h.waitingUsers[session.UserID] = true

					// Отправляем уведомление пользователю
					for client := range session.UserClients {
						h.sendSuccess(client, "Manager disconnected", nil) // ЖДИТЕ НООВОГО МЕНЕДЖЕРА ТИПА
					}

				}
			}
			delete(h.managerSessions, c.ID) // Удаляем все сессии менеджера
		}
	} else {
		// Логика отключения пользователя
		if session, ok := h.sessions[c.ID]; ok {
			session.mu.Lock()
			// Удаляем конкретное соединение из пула сессии
			delete(session.UserClients, c)
			session.mu.Unlock()

			// Если у пользователя не осталось активных соединений,
			// уведомляем менеджера.
			session.mu.RLock()
			shouldNotifyManager := len(session.UserClients) == 0 && session.Manager != nil
			session.mu.RUnlock()
			if shouldNotifyManager {
				payload, _ := json.Marshal(map[string]interface{}{
					"event":   "user_disconnected_all", // Новый, более точный event
					"user_id": c.ID,
					"message": "User has closed all connections.",
				})
				safeSend(session.Manager, payload)
			}
			// ВАЖНО: саму сессию (h.sessions[c.ID]) мы не удаляем.
			// Она хранит историю чата и может быть возобновлена.
		}
	}
}

func (h *Hub) routeMessage(sender *Client, message []byte) {
	if sender.IsManager {
		h.handleManagerMessage(sender, message) // TODO НОРМ???????????????????????????77
	} else {
		h.handleUserMessage(sender, message)
	}
}

func (h *Hub) handleUserMessage(user *Client, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, ok := h.sessions[user.ID]
	if !ok {
		logger.Warn("Session not found", map[string]interface{}{"user_id": user.ID})
		return
	}

	var input IncomingUserMessage
	if err := json.Unmarshal(message, &input); err != nil {
		input = IncomingUserMessage{
			Content: string(message),
			Type:    MsgTypeText,
		}
	}

	if input.Type == "" {
		input.Type = MsgTypeText
	}

	// 1. Создаем структуру (добавляем время!)
	msg := &Message{
		FromID:    user.ID,
		ToID:      0,
		Content:   input.Content,
		Type:      input.Type,
		FileName:  input.FileName,
		CreatedAt: time.Now(), // <--- ФИКС ВРЕМЕНИ
	}

	// Логика определения получателя
	if session.Manager != nil {
		msg.ToID = session.Manager.ID
	} else {
		if _, exists := h.waitingUsers[user.ID]; !exists {
			h.waitingUsers[user.ID] = true
		}
	}

	// 2. СНАЧАЛА СОХРАНЯЕМ (чтобы получить ID из базы)
	h.saveMessageAndAppendToHistory(session, msg)

	// 3. ТОЛЬКО ТЕПЕРЬ МАРШАЛИМ (теперь внутри msg есть реальный ID)
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal message:", err)
		return
	}

	// 4. Отправляем
	if session.Manager != nil {
		safeSend(session.Manager, data)
	}

	session.mu.RLock()
	for client := range session.UserClients {
		safeSend(client, data)
	}
	session.mu.RUnlock()
}

func (h *Hub) handleManagerMessage(manager *Client, message []byte) {
	var cmd ManagerCommand
	if err := json.Unmarshal(message, &cmd); err == nil && cmd.Command != "" {
		switch cmd.Command {
		case "take":
			h.assignManagerToUser(manager, cmd.UserID)
		case "list":
			h.listWaitingUsers(manager)
		case "close":
			h.closeSession(manager, cmd.UserID)
		}
		return
	}
	// Если не команда — значит обычное текстовое сообщение
	h.handleManagerTextMessage(manager, message)
}

func (h *Hub) handleManagerTextMessage(manager *Client, message []byte) {
	var msg ManagerMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		logger.Warn("Failed to parse manager text message", err)
		h.sendError(manager, "Invalid message format")
		return
	}

	if msg.Type == "" {
		msg.Type = MsgTypeText
	}

	h.mu.RLock() // Используем блокировку на чтение, т.к. только ищем сессию
	session := h.findSessionForManager(manager.ID, msg.UserID)
	h.mu.RUnlock()

	if session == nil {
		h.sendError(manager, "No active session with this user")
		return
	}

	// Сохраняем и отправляем сообщение с учетом типа
	messageObj := &Message{
		FromID:   manager.ID,
		ToID:     session.UserID,
		Content:  msg.Content,
		Type:     msg.Type,     // <--
		FileName: msg.FileName, // <--
	}

	h.saveMessageAndAppendToHistory(session, messageObj)
	msgData, _ := json.Marshal(messageObj)

	// Создаем локальную копию списка клиентов, чтобы не держать блокировку во время отправки
	var clientsToSend []*Client

	session.mu.RLock()
	for client := range session.UserClients {
		clientsToSend = append(clientsToSend, client)
	}
	session.mu.RUnlock() // Разблокируем сессию сразу после копирования

	// Отправляем сообщение всем активным клиентам пользователя, итерируясь по локальной копии
	for _, client := range clientsToSend {
		safeSend(client, msgData)
	}
	//session.User.Send <- msgData
}

// Модифицированная функция назначения менеджера:
func (h *Hub) assignManagerToUser(manager *Client, userID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Ищем пользователя в waitingUsers
	if _, ok := h.waitingUsers[userID]; !ok {
		h.sendError(manager, fmt.Sprintf("User %d is not waiting.", userID))
		return
	}

	// Получаем пользователя из основной сессии, которая должна быть в h.sessions
	userSession, ok := h.sessions[userID]
	if !ok {
		h.sendError(manager, fmt.Sprintf("Session for user %d not found.", userID))
		return
	}

	userSession.Manager = manager
	h.managerSessions[manager.ID] = append(h.managerSessions[manager.ID], userSession)

	delete(h.waitingUsers, userID) // Удаляем пользователя из очереди ожидания

	// Отправляем историю менеджеру
	for _, msg := range userSession.History {
		msgData, _ := json.Marshal(msg)
		safeSend(manager, msgData)
	}

	h.sendSuccess(manager, fmt.Sprintf("Session %d started with user %d", userSession.ID, userID), nil)
	userSession.mu.RLock()
	for client := range userSession.UserClients {
		h.sendSuccess(client, "A manager has joined your chat.", nil)
	}
	userSession.mu.RUnlock()
}

func (h *Hub) listWaitingUsers(manager *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.waitingUsers) == 0 {
		h.sendSuccess(manager, "No users waiting.", []uint{})
		return
	}

	var userIDs []uint
	for userID := range h.waitingUsers {
		userIDs = append(userIDs, userID)
	}

	h.sendSuccess(manager, "Waiting users list.", userIDs)
}

func (h *Hub) closeSession(manager *Client, userID uint) {
	// только отвязывает менеджера
	h.mu.Lock()
	defer h.mu.Unlock()

	sessions := h.managerSessions[manager.ID]
	for i, s := range sessions {
		if s.UserID == userID {
			h.managerSessions[manager.ID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}

	session, ok := h.sessions[userID]
	if !ok || session.Manager != manager {
		h.sendError(manager, "Cannot close this session.")
		return
	}

	// Отвязать менеджера
	session.Manager = nil

	//уже всё удаляли (но сессию оставить)
	h.sendSuccess(manager, "Session closed.", nil)
}

// ------ Новые методы для отправки JSON -----

func (h *Hub) sendResponse(client *Client, status, message string, payload interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn(" Обработать закрытый канал")
			// Обработать закрытый канал
		}
	}()
	response := ServerResponse{
		Status:  status,
		Message: message,
		Payload: payload,
	}
	respBytes, _ := json.Marshal(response)
	client.Send <- respBytes
}

func (h *Hub) sendSuccess(client *Client, message string, payload interface{}) {
	h.sendResponse(client, "success", message, payload)
}

func (h *Hub) sendError(client *Client, message string) {
	h.sendResponse(client, "error", message, nil)
}

// // Уведомление об отключении клиента
// func (h *Hub) notifyClientDisconnected(client *Client) {
// 	// Отправка уведомления всем менеджерам о том, что клиент отключился
// 	for _, session := range h.sessions {
// 		if session.Manager != nil && session.User == client {
// 			safeSend(session.Manager, []byte(fmt.Sprintf("User %d has disconnected.", client.ID)))

// 		}
// 	}
// }

// Новая функция поиска сессии:
func (h *Hub) findSessionForManager(managerID, userID uint) *Session {

	for _, session := range h.managerSessions[managerID] {
		if session.UserID == userID {
			return session
		}
	}
	return nil
}

func (h *Hub) saveMessageAndAppendToHistory(session *Session, msg *Message) {
	if err := h.ChatRepository.SaveMessage(msg); err != nil {
		logger.Error("failed to save message:", err)
	}

	// Блокируем сессию на запись перед изменением истории
	session.mu.Lock()
	session.History = append(session.History, msg)
	session.mu.Unlock()
}

func (h *Hub) listActiveSessions(manager *Client) {
	//unused (for admin mb?)
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions := h.managerSessions[manager.ID]
	response := make([]map[string]interface{}, 0, len(sessions))

	for _, s := range sessions {
		if s.UserID == 0 {
			continue
		}
		response = append(response, map[string]interface{}{
			"user_id": s.UserID,
		})
	}

	h.sendSuccess(manager, "Active sessions", response)
}

func (h *Hub) loadPreviousMessages(session *Session, limit int) { //TODO вызывается при скролле вверх , через js
	if len(session.History) == 0 {
		return // нет точек отсчета
	}

	firstMsg := session.History[0]

	// Загружаем сообщения из базы, отправленные ПЕРЕД первым в истории
	messages, err := h.ChatRepository.GetMessagesBefore(session.UserID, firstMsg.ID, limit)
	if err != nil {
		logger.Error("failed to load previous messages:", err)
		return
	}

	// Вставляем в начало истории
	session.History = append(messages, session.History...)
}

func (h *Hub) LoadLastMessages(userID uint, session *Session, limit int) error {
	messages, err := h.ChatRepository.GetLastMessages(userID, limit)
	if err != nil {
		return err
	}

	session.History = append([]*Message{}, messages...)
	return nil
}

func generateUniqueID() uint {
	return uint(atomic.AddUint64(&sessionIDCounter, 1))
}
func (h *Hub) GetSessionHistory(userID uint) []*Message {
	// be careful, no mutex

	if session, ok := h.sessions[userID]; ok {
		return session.History
	}
	return nil
}

func (h *Hub) GetSessionHistoryForManager(managerID, userID uint) []*Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, session := range h.managerSessions[managerID] {
		if session.UserID == userID {
			return session.History
		}
	}
	return nil
}

func safeSend(c *Client, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn("Recovered in safeSend: possibly sending to closed channel", map[string]interface{}{
				"client_id": c.ID,
			})
		}
	}()

	select {
	case c.Send <- data:
	default:
		// Можно залогировать или дропнуть
		logger.Warn("Send buffer full or slow client", map[string]interface{}{"client_id": c.ID})
	}
}
func sendHistoryToClient(c *Client, history []*Message) {
	for _, msg := range history {
		msgData, err := json.Marshal(msg)
		if err != nil {
			logger.Error("failed to marshal message:", err)
			continue
		}
		safeSend(c, msgData)
	}
}
