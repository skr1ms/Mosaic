package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{db: db}
}

// GetUsersByRole возвращает список пользователей по роли
func (r *Repository) GetUsersByRole(ctx context.Context, role string) ([]User, error) {
	var users []User
	
	// Для админов показываем партнеров, для партнеров - админов
	targetRole := "partner"
	if role == "partner" {
		targetRole = "admin"
	}

	// Получаем пользователей из таблицы partners или admins
	if targetRole == "partner" {
		// Получаем партнеров из базы данных
		rows, err := r.db.NewRaw(`
			SELECT 
				id::text as id,
				name,
				email,
				'partner' as role,
				true as is_online
			FROM partners 
			WHERE blocked = false
		`).Rows()
		if err != nil {
			return nil, fmt.Errorf("failed to get partners: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var user User
			err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.IsOnline)
			if err != nil {
				continue
			}
			users = append(users, user)
		}
	} else {
		// Для админов создаем тестовые данные
		users = []User{
			{ID: "admin1", Name: "Администратор 1", Email: "admin1@example.com", Role: "admin", IsOnline: true},
			{ID: "admin2", Name: "Администратор 2", Email: "admin2@example.com", Role: "admin", IsOnline: true},
		}
	}

	return users, nil
}

// GetMessages возвращает сообщения между двумя пользователями
func (r *Repository) GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error) {
	var messages []Message
	
	// Получаем сообщения из базы данных
	err := r.db.NewSelect().
		Model(&messages).
		Where("(sender_id = ? AND target_id = ?) OR (sender_id = ? AND target_id = ?)", 
			senderID, targetID, targetID, senderID).
		Order("timestamp ASC").
		Scan(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Если сообщений нет, возвращаем пустой массив
	if len(messages) == 0 {
		return []Message{}, nil
	}

	return messages, nil
}

// SaveMessage сохраняет новое сообщение
func (r *Repository) SaveMessage(ctx context.Context, message *Message) error {
	message.Timestamp = time.Now()
	_, err := r.db.NewInsert().Model(message).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

// MarkMessagesAsRead помечает сообщения как прочитанные
func (r *Repository) MarkMessagesAsRead(ctx context.Context, senderID, targetID string) error {
	_, err := r.db.NewUpdate().
		Model((*Message)(nil)).
		Set("read = true").
		Where("sender_id = ? AND target_id = ? AND read = false", senderID, targetID).
		Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}
	return nil
}

// GetUnreadCount возвращает количество непрочитанных сообщений
func (r *Repository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*Message)(nil)).
		Where("target_id = ? AND read = false", userID).
		Count(ctx)
	
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return int64(count), nil
}
