package chat

import (
	"time"

	"github.com/uptrace/bun"
)

// Message представляет сообщение в чате
type Message struct {
	bun.BaseModel `bun:"table:messages"`

	ID        uint      `json:"id" bun:"id,pk,autoincrement"`
	SenderID  string    `json:"sender_id" bun:"sender_id,notnull"`
	TargetID  string    `json:"target_id" bun:"target_id,notnull"`
	Content   string    `json:"content" bun:"content,notnull"`
	Timestamp time.Time `json:"timestamp" bun:"timestamp,notnull"`
	Read      bool      `json:"read" bun:"read,default:false"`
}

// User представляет пользователя в чате
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsOnline bool   `json:"is_online"`
}

// ChatRequest представляет запрос на отправку сообщения
type ChatRequest struct {
	TargetID string `json:"target_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// ChatResponse представляет ответ API чата
type ChatResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// UsersResponse представляет ответ со списком пользователей
type UsersResponse struct {
	Users []User `json:"users"`
}

// MessagesResponse представляет ответ со списком сообщений
type MessagesResponse struct {
	Messages []Message `json:"messages"`
}
