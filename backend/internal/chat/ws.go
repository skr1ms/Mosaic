package chat

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	websocket "github.com/gofiber/websocket/v2"
)

type ClientInfo struct {
	Conn *websocket.Conn
	Role string
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*ClientInfo
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*ClientInfo)}
}

// Attaches user connection to hub
func (h *Hub) set(userID string, c *websocket.Conn, role string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = &ClientInfo{Conn: c, Role: role}
}

// Set attaches user connection to hub (public method for tests)
func (h *Hub) Set(userID string, c any) {
	if conn, ok := c.(*websocket.Conn); ok {
		h.set(userID, conn, "test")
	}
}

// SetWithRole attaches user connection with role to hub
func (h *Hub) SetWithRole(userID string, c any, role string) {
	if conn, ok := c.(*websocket.Conn); ok {
		h.set(userID, conn, role)
	}
}

// Removes user from hub
func (h *Hub) delete(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

// Delete removes user from hub (public method for tests)
func (h *Hub) Delete(userID string) {
	h.delete(userID)
}

// Gets user connection from hub
func (h *Hub) get(userID string) *websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client := h.clients[userID]; client != nil {
		return client.Conn
	}
	return nil
}

// Sends message to specific user
func (h *Hub) sendTo(userID string, payload any) {
	conn := h.get(userID)
	if conn == nil {
		return
	}
	data, _ := json.Marshal(payload)
	_ = conn.WriteMessage(websocket.TextMessage, data)
}

// SendTo sends message to specific user
func (h *Hub) SendTo(userID string, message any) error {
	h.sendTo(userID, message)
	return nil
}

// BroadcastToRole sends message to all users with specific role
func (h *Hub) BroadcastToRole(role string, message any) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for _, client := range h.clients {
		if client != nil && client.Conn != nil {
			// Filter by role: "admin" includes both "admin" and "main_admin"
			shouldSend := false
			if role == "admin" && (client.Role == "admin" || client.Role == "main_admin") {
				shouldSend = true
			} else if client.Role == role {
				shouldSend = true
			}

			if shouldSend {
				_ = client.Conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	}
	return nil
}

// NotifyPresence notifies about user presence
func (h *Hub) NotifyPresence(userID string, online bool) {
	message := map[string]any{
		"type": "presence",
		"data": map[string]any{
			"userID": userID,
			"online": online,
		},
	}
	h.BroadcastToRole("", message)
}

// GetOnlineUsers returns list of online users
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline checks if user is online
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.clients[userID]
	return exists
}

// Socket for chat, receives messages from client and sends them to hub
func (handler *ChatHandler) Socket(c *websocket.Conn) {
	token := c.Query("token")
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = token[7:]
	}
	claims, err := handler.deps.JwtService.ValidateAccessToken(token)
	if err != nil {
		_ = c.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","data":"unauthorized"}`))
		_ = c.Close()
		return
	}

	userID := claims.UserID.String()
	userRole := claims.Role
	handler.deps.ChatService.GetHub().SetWithRole(userID, c, userRole)
	go func() {
		handler.deps.ChatService.NotifyPresence(userID, true)
	}()

	// Configure ping/pong keepalive
	c.SetReadLimit(1024 * 1024)
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Ping/pong keepalive
	go func(conn *websocket.Conn) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second)); err != nil {
				_ = conn.Close()
				return
			}
		}
	}(c)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		var env wsEnvelope
		if json.Unmarshal(msg, &env) == nil && env.Type == "message" {
			raw, _ := json.Marshal(env.Data)
			var req ChatRequest
			if json.Unmarshal(raw, &req) == nil {
				m, err := handler.deps.ChatService.SendMessage(context.Background(), userID, req.TargetID, req.Content)
				if err == nil && m != nil {
					payload := wsEnvelope{Type: "message", Data: m}
					handler.deps.ChatService.GetHub().SendTo(req.TargetID, payload)
					handler.deps.ChatService.GetHub().SendTo(userID, payload)
				}
			}
		} else if env.Type == "read" {
			// Proxy read notification to recipient
			// Expected Data: { by_user_id, user_id }
			handler.deps.ChatService.GetHub().SendTo(userID, wsEnvelope{Type: "read", Data: env.Data})
		} else if env.Type == "update" {
			// Message update by author
			raw, _ := json.Marshal(env.Data)
			var body struct {
				ID      uint   `json:"id"`
				Content string `json:"content"`
			}
			if json.Unmarshal(raw, &body) == nil && body.ID != 0 && strings.TrimSpace(body.Content) != "" {
				// Get recipient
				msg, err := handler.deps.ChatService.GetChatRepository().GetMessageByID(context.Background(), body.ID)
				if err == nil && msg.SenderID == userID {
					if err := handler.deps.ChatService.GetChatRepository().UpdateMessage(context.Background(), body.ID, userID, body.Content); err == nil {
						// Send event to both parties
						payload := wsEnvelope{Type: "message_update", Data: map[string]any{
							"id":      body.ID,
							"content": body.Content,
							"edited":  true,
						}}
						handler.deps.ChatService.GetHub().SendTo(msg.TargetID, payload)
						handler.deps.ChatService.GetHub().SendTo(userID, payload)
					}
				}
			}
		} else if env.Type == "delete" {
			// Message deletion by author
			raw, _ := json.Marshal(env.Data)
			var body struct {
				ID uint `json:"id"`
			}
			if json.Unmarshal(raw, &body) == nil && body.ID != 0 {
				// Get message to know recipient, then delete
				msg, err := handler.deps.ChatService.GetChatRepository().GetMessageByID(context.Background(), body.ID)
				if err == nil && msg.SenderID == userID {
					// Save attachment key and delete from DB by id
					attachmentKey := msg.AttachmentURL
					if err := handler.deps.ChatService.GetChatRepository().DeleteMessageByID(context.Background(), body.ID); err == nil {
						if attachmentKey != "" && handler.deps.ChatService.GetS3Client() != nil {
							_ = handler.deps.ChatService.GetS3Client().DeleteChatData(context.Background(), attachmentKey)
						}
						payload := wsEnvelope{Type: "message_delete", Data: map[string]any{"id": body.ID}}
						handler.deps.ChatService.GetHub().SendTo(msg.TargetID, payload)
						handler.deps.ChatService.GetHub().SendTo(userID, payload)
					} else {
						// Return error notification to author
						handler.deps.ChatService.GetHub().SendTo(userID, wsEnvelope{Type: "error", Data: "failed_to_delete"})
					}
				}
			}
		}
	}
	handler.deps.ChatService.GetHub().Delete(userID)
	go func() {
		handler.deps.ChatService.NotifyPresence(userID, false)
	}()
}
