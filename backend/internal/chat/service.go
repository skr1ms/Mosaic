package chat

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ChatServiceDeps struct {
	ChatRepository ChatRepositoryInterface
	S3Client       S3UploaderInterface
	Hub            HubInterface
}

type ChatService struct {
	deps *ChatServiceDeps
}

func NewChatService(deps *ChatServiceDeps) *ChatService {
	return &ChatService{
		deps: deps,
	}
}

// StartSupportChat creates new support chat with guest user
func (s *ChatService) StartSupportChat(ctx context.Context, title string, guestID uuid.UUID) (*SupportChat, error) {
	if title == "" {
		title = time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	}
	return s.deps.ChatRepository.CreateSupportChat(ctx, title, guestID)
}

// ListSupportChats retrieves all support chats
func (s *ChatService) ListSupportChats(ctx context.Context) ([]SupportChat, error) {
	return s.deps.ChatRepository.ListSupportChats(ctx)
}

// DeleteSupportChat deletes support chat with all messages and attachments
func (s *ChatService) DeleteSupportChat(ctx context.Context, chatID uuid.UUID) error {
	messages, err := s.deps.ChatRepository.GetSupportMessages(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get support messages for cleanup: %w", err)
	}

	if s.deps.S3Client != nil {
		for _, msg := range messages {
			if msg.AttachmentURL != "" {
				_ = s.deps.S3Client.DeleteChatData(ctx, msg.AttachmentURL)
			}
		}
	}

	return s.deps.ChatRepository.DeleteSupportChatCascade(ctx, chatID)
}

// GetSupportMessages retrieves all messages for support chat and marks them as read if admin
func (s *ChatService) GetSupportMessages(ctx context.Context, chatID uuid.UUID, currentUserID string, currentUserRole string) ([]SupportMessage, error) {
	messages, err := s.deps.ChatRepository.GetSupportMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// Convert S3 keys to public URLs for messages with attachments
	for i := range messages {
		if messages[i].AttachmentURL != "" && s.deps.S3Client != nil {
			messages[i].AttachmentURL = s.deps.S3Client.GetChatFileURL(messages[i].AttachmentURL)
		}
	}

	// Mark messages as read when admin views them
	if err := s.deps.ChatRepository.MarkSupportMessagesAsRead(ctx, chatID, currentUserID, currentUserRole); err != nil {
		return nil, fmt.Errorf("failed to mark support messages as read: %w", err)
	}

	// Send read notification via WebSocket if admin is viewing
	if (currentUserRole == "admin" || currentUserRole == "main_admin") && s.deps.Hub != nil {
		// Get the chat to find guest ID
		if chatRec, err := s.deps.ChatRepository.GetSupportChatByID(ctx, chatID); err == nil {
			// Notify guest that their messages were read
			payload := wsEnvelope{Type: "support_messages_read", Data: map[string]any{
				"chat_id":      chatID.String(),
				"read_by":      currentUserID,
				"read_by_role": currentUserRole,
			}}
			s.deps.Hub.SendTo(chatRec.GuestID.String(), payload)
		}
	}

	return messages, nil
}

// SendSupportMessage sends new message in support chat and broadcasts to participants
func (s *ChatService) SendSupportMessage(ctx context.Context, chatID uuid.UUID, senderID string, senderRole string, content string) (*SupportMessage, error) {
	// Allow empty content for attachment-only messages
	if strings.TrimSpace(content) == "" {
		content = ""
	}
	msg := &SupportMessage{
		ChatID:     chatID,
		SenderID:   senderID,
		SenderRole: senderRole,
		Content:    content,
	}
	if err := s.deps.ChatRepository.SaveSupportMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save support message: %w", err)
	}

	if s.deps.Hub != nil {
		s.deps.Hub.BroadcastToRole("admin", wsEnvelope{Type: "support_new_message", Data: msg})
		if chatRec, err := s.deps.ChatRepository.GetSupportChatByID(ctx, chatID); err == nil {
			_ = s.deps.Hub.SendTo(chatRec.GuestID.String(), wsEnvelope{Type: "support_new_message", Data: msg})
		}
	}
	return msg, nil
}

// UpdateSupportMessage updates support message content
func (s *ChatService) UpdateSupportMessage(ctx context.Context, id uint, senderID string, content string) error {
	return s.deps.ChatRepository.UpdateSupportMessage(ctx, id, senderID, content)
}

// DeleteSupportMessage deletes support message
func (s *ChatService) DeleteSupportMessage(ctx context.Context, id uint, senderID string) error {
	return s.deps.ChatRepository.DeleteSupportMessage(ctx, id, senderID)
}

// UploadSupportAttachment uploads attachment for support message
func (s *ChatService) UploadSupportAttachment(ctx context.Context, messageID uint, senderID string, file io.Reader, size int64, contentType, filename string) (string, error) {
	if s.deps.S3Client == nil {
		return "", fmt.Errorf("s3 is not configured")
	}
	msg, err := s.deps.ChatRepository.GetSupportMessageByID(ctx, messageID)
	if err != nil {
		return "", fmt.Errorf("failed to get support message: %w", err)
	}
	if msg.SenderID != senderID {
		return "", fmt.Errorf("forbidden")
	}
	key, publicURL, err := s.deps.S3Client.UploadChatData(ctx, file, size, contentType, senderID, msg.ChatID.String(), filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload support chat data: %w", err)
	}

	// Store the key for internal reference
	if err := s.deps.ChatRepository.SetSupportAttachmentURL(ctx, messageID, senderID, key); err != nil {
		return "", fmt.Errorf("failed to set support attachment URL: %w", err)
	}

	// Send WebSocket notification about attachment update
	if s.deps.Hub != nil {
		// Get updated message
		if updatedMsg, err := s.deps.ChatRepository.GetSupportMessageByID(ctx, messageID); err == nil {
			// Notify admin and guest about message update with attachment
			s.deps.Hub.BroadcastToRole("admin", wsEnvelope{Type: "support_message_update", Data: updatedMsg})
			if chatRec, err := s.deps.ChatRepository.GetSupportChatByID(ctx, updatedMsg.ChatID); err == nil {
				_ = s.deps.Hub.SendTo(chatRec.GuestID.String(), wsEnvelope{Type: "support_message_update", Data: updatedMsg})
			}
		}
	}

	return publicURL, nil
}

// DownloadSupportAttachment downloads support message attachment
func (s *ChatService) DownloadSupportAttachment(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	if s.deps.S3Client == nil {
		return nil, "", fmt.Errorf("s3 is not configured")
	}
	return s.deps.S3Client.DownloadChatData(ctx, objectKey)
}

// GetUsers retrieves users list for chat by role and search criteria
func (s *ChatService) GetUsers(ctx context.Context, role string, search string) ([]User, error) {
	if role == "" {
		return nil, fmt.Errorf("role is required")
	}

	users, err := s.deps.ChatRepository.GetUsersByRole(ctx, role, search)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat users: %w", err)
	}

	return users, nil
}

// GetMessages retrieves messages between two users and marks them as read
func (s *ChatService) GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error) {
	if senderID == "" || targetID == "" {
		return nil, fmt.Errorf("sender_id and target_id are required")
	}

	messages, err := s.deps.ChatRepository.GetMessages(ctx, senderID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert S3 keys to public URLs for messages with attachments
	for i := range messages {
		if messages[i].AttachmentURL != "" && s.deps.S3Client != nil {
			messages[i].AttachmentURL = s.deps.S3Client.GetChatFileURL(messages[i].AttachmentURL)
		}
	}

	if err := s.deps.ChatRepository.MarkMessagesAsRead(ctx, targetID, senderID); err != nil {
		return nil, fmt.Errorf("failed to mark as read: %w", err)
	}

	if s.deps.Hub != nil {
		payload := wsEnvelope{Type: "read", Data: map[string]any{
			"by_user_id": senderID,
			"user_id":    targetID,
		}}
		s.deps.Hub.SendTo(targetID, payload)
	}

	return messages, nil
}

// SendMessage sends new message with partner block validation
func (s *ChatService) SendMessage(ctx context.Context, senderID, targetID, content string) (*Message, error) {
	if senderID == "" || targetID == "" {
		return nil, fmt.Errorf("sender_id and target_id are required")
	}
	// Allow empty content for attachment-only messages
	if strings.TrimSpace(content) == "" {
		content = ""
	}

	if blocked, _ := s.deps.ChatRepository.IsPartnerBlocked(ctx, senderID); blocked {
		return nil, fmt.Errorf("partner is blocked")
	}

	message := &Message{
		SenderID: senderID,
		TargetID: targetID,
		Content:  content,
	}

	err := s.deps.ChatRepository.SaveMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return message, nil
}

// GetUnreadCount retrieves unread messages count for user
func (s *ChatService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, fmt.Errorf("user_id is required")
	}

	count, err := s.deps.ChatRepository.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// GetUnreadBySender retrieves unread messages count grouped by sender
func (s *ChatService) GetUnreadBySender(ctx context.Context, userID string) ([]UnreadBySender, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	rows, err := s.deps.ChatRepository.GetUnreadCountsBySender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread by sender: %w", err)
	}
	return rows, nil
}

// UploadAttachment uploads attachment and saves reference in message
func (s *ChatService) UploadAttachment(ctx context.Context, messageID uint, senderID string, file io.Reader, size int64, contentType, filename string) (string, error) {
	if s.deps.S3Client == nil {
		return "", fmt.Errorf("s3 is not configured")
	}
	msg, err := s.deps.ChatRepository.GetMessageByID(ctx, messageID)
	if err != nil {
		return "", fmt.Errorf("failed to get message: %w", err)
	}
	if msg.SenderID != senderID {
		return "", fmt.Errorf("forbidden")
	}
	key, publicURL, err := s.deps.S3Client.UploadChatData(ctx, file, size, contentType, senderID, msg.TargetID, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload chat data: %w", err)
	}

	// Store the key for internal reference, but return the public URL
	if err := s.deps.ChatRepository.SetAttachmentURL(ctx, messageID, senderID, key); err != nil {
		return "", fmt.Errorf("failed to set attachment URL: %w", err)
	}

	return publicURL, nil
}

// DownloadAttachment downloads attachment stream and content type
func (s *ChatService) DownloadAttachment(ctx context.Context, key string) (io.ReadCloser, string, error) {
	if s.deps.S3Client == nil {
		return nil, "", fmt.Errorf("s3 is not configured")
	}
	return s.deps.S3Client.DownloadChatData(ctx, key)
}

// AdminBlockPartner blocks partner in chat (admin only)
func (s *ChatService) AdminBlockPartner(ctx context.Context, partnerID string) error {
	if strings.TrimSpace(partnerID) == "" {
		return fmt.Errorf("partnerID is required")
	}
	return s.deps.ChatRepository.UpdatePartnerChatBlock(ctx, partnerID, true)
}

// AdminUnblockPartner unblocks partner in chat (admin only)
func (s *ChatService) AdminUnblockPartner(ctx context.Context, partnerID string) error {
	if strings.TrimSpace(partnerID) == "" {
		return fmt.Errorf("partnerID is required")
	}
	return s.deps.ChatRepository.UpdatePartnerChatBlock(ctx, partnerID, false)
}

// Repository access methods for WebSocket integration
func (s *ChatService) GetChatRepository() ChatRepositoryInterface {
	return s.deps.ChatRepository
}

func (s *ChatService) GetS3Client() S3UploaderInterface {
	return s.deps.S3Client
}

func (s *ChatService) GetHub() HubInterface {
	return s.deps.Hub
}

func (s *ChatService) AttachHub(hub HubInterface) {
	s.deps.Hub = hub
}

func (s *ChatService) NotifyPresence(userID string, online bool) {
	if s.deps.Hub != nil {
		s.deps.Hub.NotifyPresence(userID, online)
	}
}
