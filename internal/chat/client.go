package chat

import (
	"log"
	"time"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	ID        uint
	IsManager bool
}

func NewClient(conn *websocket.Conn, hub *Hub, userID uint, isManager bool) *Client {
	return &Client{
		Hub:       hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		ID:        userID,
		IsManager: isManager,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
		close(c.Send) // Закрываем канал Send
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {

		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Read error from client %d: %v\n", c.ID, err)
			break
		}
		c.Hub.routeMessage(c, message)
	}
}

// func (c *Client) WritePump() {
// 	ticker := time.NewTicker(pingPeriod)
// 	defer func() {
// 		ticker.Stop()
// 		c.Conn.Close()
// 	}()

// 	for {
// 		select {
// 		case message, ok := <-c.Send:
// 			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
// 			if !ok {
// 				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
// 				return
// 			}

// 			w, err := c.Conn.NextWriter(websocket.TextMessage)
// 			if err != nil {
// 				log.Printf("NextWriter error for client %d: %v\n", c.ID, err)
// 				return
// 			}
// 			w.Write(message)

// 			n := len(c.Send)
// 			for i := 0; i < n; i++ {
// 				w.Write([]byte{'\n'})
// 				w.Write(<-c.Send)
// 			}

// 			if err := w.Close(); err != nil {
// 				log.Printf("Writer close error for client %d: %v\n", c.ID, err)
// 				return
// 			}

//			case <-ticker.C:
//				c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
//				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
//					log.Printf("Ping error for client %d: %v\n", c.ID, err)
//					return
//				}
//			}
//		}
//	}
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		logger.Debug("WritePump closed", map[string]interface{}{"client_id": c.ID})
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if !ok {
				logger.Debug("Send channel closed", map[string]interface{}{"client_id": c.ID})
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage) // TODO без буферизации
			if err != nil {
				logger.Error("Failed to get writer", map[string]interface{}{
					"client_id": c.ID,
					"error":     err.Error(),
				})
				return
			}

			if _, err := w.Write(message); err != nil {
				logger.Error("Failed to write message", map[string]interface{}{
					"client_id": c.ID,
					"error":     err.Error(),
				})
				return
			}

			// Логирование отправленных сообщений
			logger.Debug("Message sent", map[string]interface{}{
				"client_id": c.ID,
				"message":   string(message),
			})

			if err := w.Close(); err != nil {
				logger.Error("Failed to close writer", map[string]interface{}{
					"client_id": c.ID,
					"error":     err.Error(),
				})
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("Failed to send ping", map[string]interface{}{
					"client_id": c.ID,
					"error":     err.Error(),
				})
				return
			}
		}
	}
}
