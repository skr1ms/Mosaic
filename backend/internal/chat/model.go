package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Message struct {
	bun.BaseModel `bun:"table:messages"`

	ID            uint      `json:"id" bun:"id,pk,autoincrement"`
	SenderID      string    `json:"sender_id" bun:"sender_id,notnull"`
	TargetID      string    `json:"target_id" bun:"target_id,notnull"`
	Content       string    `json:"content" bun:"content,notnull"`
	Timestamp     time.Time `json:"timestamp" bun:"timestamp,notnull"`
	Read          bool      `json:"read" bun:"read,default:false"`
	AttachmentURL string    `json:"attachment_url" bun:"attachment_url,nullzero"`
	Edited        bool      `json:"edited" bun:"edited,default:false"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,nullzero,default:current_timestamp"`
}

type SupportChat struct {
	bun.BaseModel `bun:"table:support_chats"`

	ID        uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Title     string    `json:"title" bun:"title,notnull"`
	GuestID   uuid.UUID `json:"guest_id" bun:"guest_id,type:uuid,notnull"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
}

type SupportMessage struct {
	bun.BaseModel `bun:"table:support_messages"`

	ID            uint      `json:"id" bun:"id,pk,autoincrement"`
	ChatID        uuid.UUID `json:"chat_id" bun:"chat_id,type:uuid,notnull"`
	SenderID      string    `json:"sender_id" bun:"sender_id,notnull"`
	SenderRole    string    `json:"sender_role" bun:"sender_role,notnull"`
	Content       string    `json:"content" bun:"content,notnull"`
	AttachmentURL string    `json:"attachment_url" bun:"attachment_url,nullzero"`
	Edited        bool      `json:"edited" bun:"edited,default:false"`
	Read          bool      `json:"read" bun:"read,default:false"`
	Timestamp     time.Time `json:"timestamp" bun:"timestamp,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `json:"updated_at" bun:"updated_at,nullzero,default:current_timestamp"`
}

type MessageResponse struct {
	ID            uint      `json:"id"`
	SenderID      string    `json:"sender_id"`
	TargetID      string    `json:"target_id"`
	Content       string    `json:"content"`
	Timestamp     time.Time `json:"timestamp"`
	Read          bool      `json:"read"`
	AttachmentURL string    `json:"attachment_url"`
	Edited        bool      `json:"edited"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	IsOnline        bool   `json:"is_online"`
	Status          string `json:"status,omitempty"`
	PartnerCode     string `json:"partner_code,omitempty"`
	Login           string `json:"login,omitempty"`
	IsBlockedInChat bool   `json:"is_blocked_in_chat,omitempty"`
}

type ChatRequest struct {
	TargetID string `json:"target_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

type ChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type UsersResponse struct {
	Users []User `json:"users"`
}

type MessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}
