package chat

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type ChatServiceInterface interface {
	GetUsers(ctx context.Context, role string, search string) ([]User, error)
	GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error)
	SendMessage(ctx context.Context, senderID, targetID, content string) (*Message, error)
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
	GetUnreadBySender(ctx context.Context, userID string) ([]UnreadBySender, error)
	UploadAttachment(ctx context.Context, messageID uint, senderID string, file io.Reader, size int64, contentType, filename string) (string, error)
	DownloadAttachment(ctx context.Context, key string) (io.ReadCloser, string, error)
	AdminBlockPartner(ctx context.Context, partnerID string) error
	AdminUnblockPartner(ctx context.Context, partnerID string) error
	AttachHub(hub HubInterface)

	GetChatRepository() ChatRepositoryInterface
	GetS3Client() S3UploaderInterface

	GetHub() HubInterface
	NotifyPresence(userID string, online bool)

	// Support chat service API
	StartSupportChat(ctx context.Context, title string, guestID uuid.UUID) (*SupportChat, error)
	ListSupportChats(ctx context.Context) ([]SupportChat, error)
	DeleteSupportChat(ctx context.Context, chatID uuid.UUID) error
	GetSupportMessages(ctx context.Context, chatID uuid.UUID, currentUserID string, currentUserRole string) ([]SupportMessage, error)
	SendSupportMessage(ctx context.Context, chatID uuid.UUID, senderID string, senderRole string, content string) (*SupportMessage, error)
	UpdateSupportMessage(ctx context.Context, id uint, senderID string, content string) error
	DeleteSupportMessage(ctx context.Context, id uint, senderID string) error

	UploadSupportAttachment(ctx context.Context, messageID uint, senderID string, file io.Reader, size int64, contentType, filename string) (string, error)
	DownloadSupportAttachment(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
}

type ChatRepositoryInterface interface {
	GetUsersByRole(ctx context.Context, role string, search string) ([]User, error)
	GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error)
	SaveMessage(ctx context.Context, message *Message) error
	MarkMessagesAsRead(ctx context.Context, senderID, targetID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
	GetMessageByID(ctx context.Context, id uint) (*Message, error)
	UpdateMessage(ctx context.Context, id uint, senderID string, content string) error
	DeleteMessage(ctx context.Context, id uint, senderID string) error
	DeleteMessageByID(ctx context.Context, id uint) error
	GetAttachmentKeyByID(ctx context.Context, id uint, senderID string) (string, error)
	SetAttachmentURL(ctx context.Context, id uint, senderID string, url string) error
	GetUnreadSenders(ctx context.Context, userID string) ([]string, error)
	GetUnreadCountsBySender(ctx context.Context, userID string) ([]UnreadBySender, error)
	IsPartnerBlocked(ctx context.Context, partnerID string) (bool, error)
	UpdatePartnerChatBlock(ctx context.Context, partnerID string, blocked bool) error
	ListAttachmentsByUser(ctx context.Context, userID string) ([]string, error)
	DeleteAllMessagesByUser(ctx context.Context, userID string) error

	// Support chat repository API
	CreateSupportChat(ctx context.Context, title string, guestID uuid.UUID) (*SupportChat, error)
	ListSupportChats(ctx context.Context) ([]SupportChat, error)
	DeleteSupportChatCascade(ctx context.Context, chatID uuid.UUID) error
	GetSupportChatByID(ctx context.Context, chatID uuid.UUID) (*SupportChat, error)
	GetSupportMessages(ctx context.Context, chatID uuid.UUID) ([]SupportMessage, error)
	SaveSupportMessage(ctx context.Context, msg *SupportMessage) error
	UpdateSupportMessage(ctx context.Context, id uint, senderID string, content string) error
	DeleteSupportMessage(ctx context.Context, id uint, senderID string) error

	GetSupportMessageByID(ctx context.Context, id uint) (*SupportMessage, error)
	SetSupportAttachmentURL(ctx context.Context, id uint, senderID string, url string) error
	GetSupportAttachmentKeyByID(ctx context.Context, id uint, senderID string) (string, error)
	MarkSupportMessagesAsRead(ctx context.Context, chatID uuid.UUID, currentUserID string, currentUserRole string) error
	GetUnreadSupportMessagesCount(ctx context.Context) (int64, error)
}

type S3UploaderInterface interface {
	UploadChatData(ctx context.Context, reader io.Reader, size int64, contentType string, senderID string, targetID string, originalFilename string) (string, string, error)
	DownloadChatData(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
	DeleteChatData(ctx context.Context, objectKey string) error
	GetChatFileURL(objectKey string) string
}

type JWTServiceInterface interface {
	CreateAccessToken(userID uuid.UUID, login, role string) (string, error)
	CreateTokenPair(userID uuid.UUID, login, role string) (*jwt.TokenPair, error)
	ValidateRefreshToken(refreshToken string) (*jwt.Claims, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
	CreatePasswordResetToken(userID uuid.UUID, email string) (string, error)
	ValidatePasswordResetToken(token string) (*jwt.Claims, error)
	ValidateAccessToken(token string) (*jwt.Claims, error)
}

type HubInterface interface {
	SendTo(userID string, message any) error
	BroadcastToRole(role string, message any) error
	NotifyPresence(userID string, online bool)
	GetOnlineUsers() []string
	IsUserOnline(userID string) bool
	Set(userID string, c any)
	SetWithRole(userID string, c any, role string)
	Delete(userID string)
}
