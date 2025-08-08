package chat

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetUsers возвращает список пользователей для чата
func (s *Service) GetUsers(ctx context.Context, role string) ([]User, error) {
	if role == "" {
		return nil, errors.New("role is required")
	}

	users, err := s.repo.GetUsersByRole(ctx, role)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetMessages возвращает сообщения между двумя пользователями
func (s *Service) GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error) {
	if senderID == "" || targetID == "" {
		return nil, errors.New("sender_id and target_id are required")
	}

	messages, err := s.repo.GetMessages(ctx, senderID, targetID)
	if err != nil {
		return nil, err
	}

	// Помечаем сообщения как прочитанные
	go func() {
		ctx := context.Background()
		s.repo.MarkMessagesAsRead(ctx, senderID, targetID)
	}()

	return messages, nil
}

// SendMessage отправляет новое сообщение
func (s *Service) SendMessage(ctx context.Context, senderID, targetID, content string) (*Message, error) {
	if senderID == "" || targetID == "" || content == "" {
		return nil, errors.New("sender_id, target_id and content are required")
	}

	message := &Message{
		SenderID: senderID,
		TargetID: targetID,
		Content:  content,
	}

	err := s.repo.SaveMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetUnreadCount возвращает количество непрочитанных сообщений для пользователя
func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, errors.New("user_id is required")
	}

	count, err := s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
