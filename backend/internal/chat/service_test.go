package chat

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) GetUsersByRole(ctx context.Context, role string, search string) ([]User, error) {
	args := m.Called(ctx, role, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error) {
	args := m.Called(ctx, senderID, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Message), args.Error(1)
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, message *Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockChatRepository) MarkMessagesAsRead(ctx context.Context, senderID, targetID string) error {
	args := m.Called(ctx, senderID, targetID)
	return args.Error(0)
}

func (m *MockChatRepository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockChatRepository) GetMessageByID(ctx context.Context, id uint) (*Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Message), args.Error(1)
}

func (m *MockChatRepository) UpdateMessage(ctx context.Context, id uint, senderID string, content string) error {
	args := m.Called(ctx, id, senderID, content)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteMessage(ctx context.Context, id uint, senderID string) error {
	args := m.Called(ctx, id, senderID)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteMessageByID(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChatRepository) GetAttachmentKeyByID(ctx context.Context, id uint, senderID string) (string, error) {
	args := m.Called(ctx, id, senderID)
	return args.String(0), args.Error(1)
}

func (m *MockChatRepository) SetAttachmentURL(ctx context.Context, id uint, senderID string, url string) error {
	args := m.Called(ctx, id, senderID, url)
	return args.Error(0)
}

func (m *MockChatRepository) GetUnreadSenders(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockChatRepository) GetUnreadCountsBySender(ctx context.Context, userID string) ([]UnreadBySender, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]UnreadBySender), args.Error(1)
}

func (m *MockChatRepository) IsPartnerBlocked(ctx context.Context, partnerID string) (bool, error) {
	args := m.Called(ctx, partnerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) UpdatePartnerChatBlock(ctx context.Context, partnerID string, blocked bool) error {
	args := m.Called(ctx, partnerID, blocked)
	return args.Error(0)
}

func (m *MockChatRepository) ListAttachmentsByUser(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockChatRepository) DeleteAllMessagesByUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Support chat mock methods
func (m *MockChatRepository) CreateSupportChat(ctx context.Context, title string, guestID uuid.UUID) (*SupportChat, error) {
	args := m.Called(ctx, title, guestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SupportChat), args.Error(1)
}

func (m *MockChatRepository) ListSupportChats(ctx context.Context) ([]SupportChat, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SupportChat), args.Error(1)
}

func (m *MockChatRepository) DeleteSupportChatCascade(ctx context.Context, chatID uuid.UUID) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockChatRepository) GetSupportMessages(ctx context.Context, chatID uuid.UUID) ([]SupportMessage, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SupportMessage), args.Error(1)
}

func (m *MockChatRepository) SaveSupportMessage(ctx context.Context, msg *SupportMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatRepository) UpdateSupportMessage(ctx context.Context, id uint, senderID string, content string) error {
	args := m.Called(ctx, id, senderID, content)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteSupportMessage(ctx context.Context, id uint, senderID string) error {
	args := m.Called(ctx, id, senderID)
	return args.Error(0)
}

func (m *MockChatRepository) GetSupportChatByID(ctx context.Context, chatID uuid.UUID) (*SupportChat, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SupportChat), args.Error(1)
}

func (m *MockChatRepository) GetSupportAttachmentKeyByID(ctx context.Context, id uint, senderID string) (string, error) {
	args := m.Called(ctx, id, senderID)
	return args.String(0), args.Error(1)
}

func (m *MockChatRepository) SetSupportAttachmentURL(ctx context.Context, id uint, senderID string, url string) error {
	args := m.Called(ctx, id, senderID, url)
	return args.Error(0)
}

func (m *MockChatRepository) GetSupportMessageByID(ctx context.Context, id uint) (*SupportMessage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SupportMessage), args.Error(1)
}

func (m *MockChatRepository) MarkSupportMessagesAsRead(ctx context.Context, chatID uuid.UUID, currentUserID string, currentUserRole string) error {
	args := m.Called(ctx, chatID, currentUserID, currentUserRole)
	return args.Error(0)
}

func (m *MockChatRepository) GetUnreadSupportMessagesCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Mock S3 Uploader
type MockS3Uploader struct {
	mock.Mock
}

func (m *MockS3Uploader) UploadChatData(ctx context.Context, reader io.Reader, size int64, contentType string, senderID string, targetID string, originalFilename string) (string, string, error) {
	args := m.Called(ctx, reader, size, contentType, senderID, targetID, originalFilename)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockS3Uploader) DownloadChatData(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	args := m.Called(ctx, objectKey)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.String(1), args.Error(2)
}

func (m *MockS3Uploader) DeleteChatData(ctx context.Context, objectKey string) error {
	args := m.Called(ctx, objectKey)
	return args.Error(0)
}

func (m *MockS3Uploader) GetChatFileURL(objectKey string) string {
	args := m.Called(objectKey)
	return args.String(0)
}

// Mock Hub
type MockHub struct {
	mock.Mock
}

func (m *MockHub) SendTo(userID string, message any) error {
	args := m.Called(userID, message)
	return args.Error(0)
}

func (m *MockHub) BroadcastToRole(role string, message any) error {
	args := m.Called(role, message)
	return args.Error(0)
}

func (m *MockHub) NotifyPresence(userID string, online bool) {
	m.Called(userID, online)
}

func (m *MockHub) GetOnlineUsers() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func (m *MockHub) IsUserOnline(userID string) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

func (m *MockHub) Set(userID string, c any) {
	m.Called(userID, c)
}

func (m *MockHub) SetWithRole(userID string, c any, role string) {
	m.Called(userID, c, role)
}

func (m *MockHub) Delete(userID string) {
	m.Called(userID)
}

func TestService_GetUsers(t *testing.T) {
	tests := []struct {
		name          string
		role          string
		search        string
		mockSetup     func(*MockChatRepository)
		expectedError bool
	}{
		{
			name:   "successful_get_users_admin",
			role:   "admin",
			search: "test",
			mockSetup: func(repo *MockChatRepository) {
				users := []User{
					{ID: "1", Name: "Admin User", Role: "admin", Email: "admin@test.com"},
				}
				repo.On("GetUsersByRole", mock.Anything, "admin", "test").Return(users, nil)
			},
			expectedError: false,
		},
		{
			name:   "successful_get_users_partner",
			role:   "partner",
			search: "",
			mockSetup: func(repo *MockChatRepository) {
				users := []User{
					{ID: "1", Name: "Partner User", Role: "partner", Email: "partner@test.com"},
				}
				repo.On("GetUsersByRole", mock.Anything, "partner", "").Return(users, nil)
			},
			expectedError: false,
		},
		{
			name:          "empty_role",
			role:          "",
			search:        "test",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:   "repository_error",
			role:   "admin",
			search: "test",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("GetUsersByRole", mock.Anything, "admin", "test").Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			result, err := service.GetUsers(context.Background(), tt.role, tt.search)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_GetMessages(t *testing.T) {
	tests := []struct {
		name          string
		senderID      string
		targetID      string
		mockSetup     func(*MockChatRepository)
		expectedError bool
	}{
		{
			name:     "successful_get_messages",
			senderID: "sender1",
			targetID: "target1",
			mockSetup: func(repo *MockChatRepository) {
				messages := []Message{
					{
						ID:        1,
						SenderID:  "sender1",
						TargetID:  "target1",
						Content:   "Hello",
						Timestamp: time.Now(),
						Read:      false,
					},
				}
				repo.On("GetMessages", mock.Anything, "sender1", "target1").Return(messages, nil)
				repo.On("MarkMessagesAsRead", mock.Anything, "target1", "sender1").Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "empty_sender_id",
			senderID:      "",
			targetID:      "target1",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:          "empty_target_id",
			senderID:      "sender1",
			targetID:      "",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:     "repository_error",
			senderID: "sender1",
			targetID: "target1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("GetMessages", mock.Anything, "sender1", "target1").Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			result, err := service.GetMessages(context.Background(), tt.senderID, tt.targetID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_SendMessage(t *testing.T) {
	tests := []struct {
		name          string
		senderID      string
		targetID      string
		content       string
		mockSetup     func(*MockChatRepository)
		expectedError bool
	}{
		{
			name:     "successful_send_message",
			senderID: "sender1",
			targetID: "target1",
			content:  "Hello, world!",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("IsPartnerBlocked", mock.Anything, "sender1").Return(false, nil)
				repo.On("SaveMessage", mock.Anything, mock.AnythingOfType("*chat.Message")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "empty_sender_id",
			senderID:      "",
			targetID:      "target1",
			content:       "Hello",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:          "empty_target_id",
			senderID:      "sender1",
			targetID:      "",
			content:       "Hello",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:     "empty_content",
			senderID: "sender1",
			targetID: "target1",
			content:  "",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("IsPartnerBlocked", mock.Anything, "sender1").Return(false, nil)
				repo.On("SaveMessage", mock.Anything, mock.AnythingOfType("*chat.Message")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "partner_blocked",
			senderID: "sender1",
			targetID: "target1",
			content:  "Hello",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("IsPartnerBlocked", mock.Anything, "sender1").Return(true, nil)
			},
			expectedError: true,
		},
		{
			name:     "repository_error",
			senderID: "sender1",
			targetID: "target1",
			content:  "Hello",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("IsPartnerBlocked", mock.Anything, "sender1").Return(false, nil)
				repo.On("SaveMessage", mock.Anything, mock.AnythingOfType("*chat.Message")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			result, err := service.SendMessage(context.Background(), tt.senderID, tt.targetID, tt.content)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_GetUnreadCount(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockChatRepository)
		expectedCount int64
		expectedError bool
	}{
		{
			name:   "successful_get_unread_count",
			userID: "user1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("GetUnreadCount", mock.Anything, "user1").Return(int64(5), nil)
			},
			expectedCount: 5,
			expectedError: false,
		},
		{
			name:          "empty_user_id",
			userID:        "",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:   "repository_error",
			userID: "user1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("GetUnreadCount", mock.Anything, "user1").Return(int64(0), errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			result, err := service.GetUnreadCount(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, result)
			}
		})
	}
}

func TestService_UploadAttachment(t *testing.T) {
	tests := []struct {
		name          string
		messageID     uint
		senderID      string
		contentType   string
		filename      string
		mockSetup     func(*MockChatRepository, *MockS3Uploader)
		expectedError bool
	}{
		{
			name:        "successful_upload",
			messageID:   1,
			senderID:    "sender1",
			contentType: "image/jpeg",
			filename:    "test.jpg",
			mockSetup: func(repo *MockChatRepository, s3 *MockS3Uploader) {
				message := &Message{
					ID:       1,
					SenderID: "sender1",
					TargetID: "target1",
					Content:  "Test message",
				}
				repo.On("GetMessageByID", mock.Anything, uint(1)).Return(message, nil)
				s3.On("UploadChatData", mock.Anything, mock.Anything, mock.Anything, "image/jpeg", "sender1", "target1", "test.jpg").Return("s3-key", "https://s3.url/file", nil)
				repo.On("SetAttachmentURL", mock.Anything, uint(1), "sender1", "s3-key").Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "zero_message_id",
			messageID:   0,
			senderID:    "sender1",
			contentType: "image/jpeg",
			filename:    "test.jpg",
			mockSetup: func(repo *MockChatRepository, s3 *MockS3Uploader) {
				repo.On("GetMessageByID", mock.Anything, uint(0)).Return(nil, errors.New("message not found"))
			},
			expectedError: true,
		},
		{
			name:        "empty_sender_id",
			messageID:   1,
			senderID:    "",
			contentType: "image/jpeg",
			filename:    "test.jpg",
			mockSetup: func(repo *MockChatRepository, s3 *MockS3Uploader) {
				message := &Message{
					ID:       1,
					SenderID: "sender1",
					TargetID: "target1",
					Content:  "Test message",
				}
				repo.On("GetMessageByID", mock.Anything, uint(1)).Return(message, nil)
			},
			expectedError: true,
		},
		{
			name:        "message_not_found",
			messageID:   1,
			senderID:    "sender1",
			contentType: "image/jpeg",
			filename:    "test.jpg",
			mockSetup: func(repo *MockChatRepository, s3 *MockS3Uploader) {
				repo.On("GetMessageByID", mock.Anything, uint(1)).Return(nil, errors.New("message not found"))
			},
			expectedError: true,
		},
		{
			name:        "forbidden_sender",
			messageID:   1,
			senderID:    "sender1",
			contentType: "image/jpeg",
			filename:    "test.jpg",
			mockSetup: func(repo *MockChatRepository, s3 *MockS3Uploader) {
				message := &Message{
					ID:       1,
					SenderID: "different_sender",
					TargetID: "target1",
					Content:  "Test message",
				}
				repo.On("GetMessageByID", mock.Anything, uint(1)).Return(message, nil)
			},
			expectedError: true,
		},
		{
			name:        "s3_upload_error",
			messageID:   1,
			senderID:    "sender1",
			contentType: "image/jpeg",
			filename:    "test.jpg",
			mockSetup: func(repo *MockChatRepository, s3 *MockS3Uploader) {
				message := &Message{
					ID:       1,
					SenderID: "sender1",
					TargetID: "target1",
					Content:  "Test message",
				}
				repo.On("GetMessageByID", mock.Anything, uint(1)).Return(message, nil)
				s3.On("UploadChatData", mock.Anything, mock.Anything, mock.Anything, "image/jpeg", "sender1", "target1", "test.jpg").Return("", "", errors.New("s3 error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockS3)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			reader := strings.NewReader("test file content")
			result, err := service.UploadAttachment(context.Background(), tt.messageID, tt.senderID, reader, 17, tt.contentType, tt.filename)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func TestService_AdminBlockPartner(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     string
		mockSetup     func(*MockChatRepository)
		expectedError bool
	}{
		{
			name:      "successful_block",
			partnerID: "partner1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("UpdatePartnerChatBlock", mock.Anything, "partner1", true).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "empty_partner_id",
			partnerID:     "",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:      "repository_error",
			partnerID: "partner1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("UpdatePartnerChatBlock", mock.Anything, "partner1", true).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			err := service.AdminBlockPartner(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetUnreadBySender(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockChatRepository)
		expectedError bool
	}{
		{
			name:   "successful_get_unread_by_sender",
			userID: "user1",
			mockSetup: func(repo *MockChatRepository) {
				unreadBySender := []UnreadBySender{
					{SenderID: "sender1", Count: 3},
					{SenderID: "sender2", Count: 2},
				}
				repo.On("GetUnreadCountsBySender", mock.Anything, "user1").Return(unreadBySender, nil)
			},
			expectedError: false,
		},
		{
			name:          "empty_user_id",
			userID:        "",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:   "repository_error",
			userID: "user1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("GetUnreadCountsBySender", mock.Anything, "user1").Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			result, err := service.GetUnreadBySender(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_DownloadAttachment(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		mockSetup     func(*MockS3Uploader)
		expectedError bool
	}{
		{
			name: "successful_download",
			key:  "test-key",
			mockSetup: func(s3 *MockS3Uploader) {
				reader := io.NopCloser(strings.NewReader("test content"))
				s3.On("DownloadChatData", mock.Anything, "test-key").Return(reader, "image/jpeg", nil)
			},
			expectedError: false,
		},
		{
			name: "empty_key",
			key:  "",
			mockSetup: func(s3 *MockS3Uploader) {
				s3.On("DownloadChatData", mock.Anything, "").Return(nil, "", errors.New("empty key"))
			},
			expectedError: true,
		},
		{
			name: "s3_download_error",
			key:  "test-key",
			mockSetup: func(s3 *MockS3Uploader) {
				s3.On("DownloadChatData", mock.Anything, "test-key").Return(nil, "", errors.New("s3 error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockS3)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			reader, contentType, err := service.DownloadAttachment(context.Background(), tt.key)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, reader)
				assert.Empty(t, contentType)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reader)
				assert.NotEmpty(t, contentType)
				if reader != nil {
					reader.Close()
				}
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func TestService_AdminUnblockPartner(t *testing.T) {
	tests := []struct {
		name          string
		partnerID     string
		mockSetup     func(*MockChatRepository)
		expectedError bool
	}{
		{
			name:      "successful_unblock",
			partnerID: "partner1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("UpdatePartnerChatBlock", mock.Anything, "partner1", false).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "empty_partner_id",
			partnerID:     "",
			mockSetup:     func(repo *MockChatRepository) {},
			expectedError: true,
		},
		{
			name:      "repository_error",
			partnerID: "partner1",
			mockSetup: func(repo *MockChatRepository) {
				repo.On("UpdatePartnerChatBlock", mock.Anything, "partner1", false).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
				},
			}

			err := service.AdminUnblockPartner(context.Background(), tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetterMethods(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockS3 := new(MockS3Uploader)
	mockHub := new(MockHub)

	service := &ChatService{
		deps: &ChatServiceDeps{
			ChatRepository: mockRepo,
			S3Client:       mockS3,
			Hub:            mockHub,
		},
	}

	// Test GetChatRepository
	assert.Equal(t, mockRepo, service.GetChatRepository())

	// Test GetS3Client
	assert.Equal(t, mockS3, service.GetS3Client())

	// Test GetHub
	assert.Equal(t, mockHub, service.GetHub())
}

func TestService_NotifyPresence(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		online    bool
		hasHub    bool
		mockSetup func(*MockHub)
	}{
		{
			name:   "notify_online_with_hub",
			userID: "user1",
			online: true,
			hasHub: true,
			mockSetup: func(mockHub *MockHub) {
				mockHub.On("NotifyPresence", "user1", true).Return()
			},
		},
		{
			name:   "notify_offline_with_hub",
			userID: "user1",
			online: false,
			hasHub: true,
			mockSetup: func(mockHub *MockHub) {
				mockHub.On("NotifyPresence", "user1", false).Return()
			},
		},
		{
			name:      "notify_without_hub",
			userID:    "user1",
			online:    true,
			hasHub:    false,
			mockSetup: func(mockHub *MockHub) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockChatRepository)
			mockS3 := new(MockS3Uploader)

			var hub HubInterface
			if tt.hasHub {
				mockHub := new(MockHub)
				if tt.mockSetup != nil {
					tt.mockSetup(mockHub)
				}
				hub = mockHub
			}

			service := &ChatService{
				deps: &ChatServiceDeps{
					ChatRepository: mockRepo,
					S3Client:       mockS3,
					Hub:            hub,
				},
			}

			// This should not panic
			assert.NotPanics(t, func() {
				service.NotifyPresence(tt.userID, tt.online)
			})
		})
	}
}

func TestService_AttachHub(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockS3 := new(MockS3Uploader)

	service := &ChatService{
		deps: &ChatServiceDeps{
			ChatRepository: mockRepo,
			S3Client:       mockS3,
			Hub:            nil,
		},
	}

	newHub := new(MockHub)
	service.AttachHub(newHub)

	assert.Equal(t, newHub, service.GetHub())
}

// Test service initialization
func TestNewChatService(t *testing.T) {
	repo := &Repository{}
	mockS3 := new(MockS3Uploader)

	deps := &ChatServiceDeps{
		ChatRepository: repo,
		S3Client:       mockS3,
		Hub:            nil,
	}

	service := NewChatService(deps)

	assert.NotNil(t, service)
	assert.NotNil(t, service.deps)
	assert.NotNil(t, service.deps.ChatRepository)
	assert.NotNil(t, service.deps.S3Client)
	assert.Equal(t, repo, service.deps.ChatRepository)
	assert.Equal(t, mockS3, service.deps.S3Client)
}
