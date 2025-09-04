package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUsersByRole(ctx context.Context, role string, search string) ([]User, error) {
	var users []User
	search = strings.TrimSpace(search)

	if role == "admin" || role == "main_admin" {
		q := r.db.NewSelect().
			TableExpr("partners AS p").
			ColumnExpr("p.id::text AS id").
			ColumnExpr("p.brand_name AS name").
			ColumnExpr("p.email AS email").
			ColumnExpr("'partner' AS role").
			ColumnExpr("false AS is_online").
			ColumnExpr("p.status AS status").
			ColumnExpr("p.partner_code AS partner_code").
			ColumnExpr("p.login AS login").
			ColumnExpr("p.is_blocked_in_chat AS is_blocked_in_chat")
		if search != "" {
			like := "%" + search + "%"
			q = q.Where("p.partner_code ILIKE ? OR p.brand_name ILIKE ? OR p.email ILIKE ?", like, like, like)
		}
		err := q.Scan(ctx, &users)
		if err != nil {
			return nil, fmt.Errorf("failed to get partners: %w", err)
		}
	} else {
		q := r.db.NewSelect().
			TableExpr("admins AS a").
			ColumnExpr("a.id::text AS id").
			ColumnExpr("a.login AS name").
			ColumnExpr("a.email AS email").
			ColumnExpr("'admin' AS role").
			ColumnExpr("false AS is_online").
			ColumnExpr("'' AS partner_code").
			ColumnExpr("a.login AS login")
		if search != "" {
			like := "%" + search + "%"
			q = q.Where("a.email ILIKE ? OR a.login ILIKE ?", like, like)
		}
		err := q.Scan(ctx, &users)
		if err != nil {
			return nil, fmt.Errorf("failed to get admins: %w", err)
		}
	}

	return users, nil
}

func (r *Repository) GetMessages(ctx context.Context, senderID, targetID string) ([]Message, error) {
	var messages []Message

	err := r.db.NewSelect().
		Model(&messages).
		Where("(sender_id = ? AND target_id = ?) OR (sender_id = ? AND target_id = ?)",
			senderID, targetID, targetID, senderID).
		Order("timestamp ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return []Message{}, nil
	}

	return messages, nil
}

func (r *Repository) SaveMessage(ctx context.Context, message *Message) error {
	message.Timestamp = time.Now()
	_, err := r.db.NewInsert().Model(message).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

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

func (r *Repository) GetMessageByID(ctx context.Context, id uint) (*Message, error) {
	msg := new(Message)
	err := r.db.NewSelect().Model(msg).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return msg, nil
}

func (r *Repository) UpdateMessage(ctx context.Context, id uint, senderID string, content string) error {
	_, err := r.db.NewUpdate().
		Model((*Message)(nil)).
		Set("content = ?", content).
		Set("edited = TRUE").
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND sender_id = ?", id, senderID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

func (r *Repository) DeleteMessage(ctx context.Context, id uint, senderID string) error {
	res, err := r.db.NewDelete().
		Model((*Message)(nil)).
		Where("id = ? AND sender_id = ?", id, senderID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("no rows deleted")
	}
	return nil
}

func (r *Repository) DeleteMessageByID(ctx context.Context, id uint) error {
	res, err := r.db.NewDelete().
		Model((*Message)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete message by id: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("no rows deleted")
	}
	return nil
}

func (r *Repository) GetAttachmentKeyByID(ctx context.Context, id uint, senderID string) (string, error) {
	var key string
	err := r.db.NewSelect().
		TableExpr("messages AS m").
		ColumnExpr("m.attachment_url").
		Where("m.id = ? AND m.sender_id = ?", id, senderID).
		Scan(ctx, &key)
	if err != nil {
		return "", fmt.Errorf("failed to get attachment key: %w", err)
	}
	return key, nil
}

func (r *Repository) SetAttachmentURL(ctx context.Context, id uint, senderID string, url string) error {
	_, err := r.db.NewUpdate().
		Model((*Message)(nil)).
		Set("attachment_url = ?", url).
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND sender_id = ?", id, senderID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to attach file to message: %w", err)
	}
	return nil
}

func (r *Repository) GetUnreadSenders(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.NewSelect().
		TableExpr("messages AS m").
		ColumnExpr("DISTINCT m.sender_id").
		Where("m.target_id = ? AND m.read = FALSE", userID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread senders: %w", err)
	}
	return ids, nil
}

type UnreadBySender struct {
	SenderID string `json:"sender_id"`
	Count    int    `json:"count"`
}

func (r *Repository) GetUnreadCountsBySender(ctx context.Context, userID string) ([]UnreadBySender, error) {
	var rows []UnreadBySender
	err := r.db.NewSelect().
		TableExpr("messages AS m").
		ColumnExpr("m.sender_id AS sender_id").
		ColumnExpr("COUNT(*) AS count").
		Where("m.target_id = ? AND m.read = FALSE", userID).
		GroupExpr("m.sender_id").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread counts by sender: %w", err)
	}
	return rows, nil
}

func (r *Repository) IsPartnerBlocked(ctx context.Context, partnerID string) (bool, error) {
	partnerID = strings.TrimSpace(partnerID)
	if partnerID == "" {
		return false, nil
	}
	var blockedInChat bool
	// If record not found - consider it's not a partner
	err := r.db.NewSelect().
		TableExpr("partners AS p").
		ColumnExpr("p.is_blocked_in_chat").
		Where("p.id::text = ?", partnerID).
		Scan(ctx, &blockedInChat)
	if err != nil {
		return false, nil
	}
	return blockedInChat, nil
}

func (r *Repository) UpdatePartnerChatBlock(ctx context.Context, partnerID string, blocked bool) error {
	partnerID = strings.TrimSpace(partnerID)
	if partnerID == "" {
		return fmt.Errorf("partnerID is required")
	}
	_, err := r.db.NewUpdate().
		TableExpr("partners").
		Set("is_blocked_in_chat = ?", blocked).
		Where("id::text = ?", partnerID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update partner chat block: %w", err)
	}
	return nil
}

func (r *Repository) ListAttachmentsByUser(ctx context.Context, userID string) ([]string, error) {
	var keys []string
	err := r.db.NewSelect().
		TableExpr("messages AS m").
		ColumnExpr("m.attachment_url").
		Where("(m.sender_id = ? OR m.target_id = ?) AND m.attachment_url IS NOT NULL AND m.attachment_url <> ''", userID, userID).
		Scan(ctx, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to list attachments: %w", err)
	}
	return keys, nil
}

func (r *Repository) DeleteAllMessagesByUser(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().
		Model((*Message)(nil)).
		Where("sender_id = ? OR target_id = ?", userID, userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete all messages: %w", err)
	}
	return nil
}

func (r *Repository) CreateSupportChat(ctx context.Context, title string, guestID uuid.UUID) (*SupportChat, error) {
	// If guest already has a chat - return the newest one to avoid duplicates
	var existing SupportChat
	if err := r.db.NewSelect().Model(&existing).Where("guest_id = ?", guestID).Order("created_at DESC").Limit(1).Scan(ctx); err == nil && existing.ID != uuid.Nil {
		return &existing, nil
	}
	chat := &SupportChat{Title: title, GuestID: guestID}
	if _, err := r.db.NewInsert().Model(chat).Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to create support chat: %w", err)
	}
	return chat, nil
}

func (r *Repository) ListSupportChats(ctx context.Context) ([]SupportChat, error) {
	var chats []SupportChat
	err := r.db.NewSelect().Model(&chats).Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list support chats: %w", err)
	}
	return chats, nil
}

func (r *Repository) DeleteSupportChatCascade(ctx context.Context, chatID uuid.UUID) error {
	// Delete messages first
	if _, err := r.db.NewDelete().Model((*SupportMessage)(nil)).Where("chat_id = ?", chatID).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete support messages: %w", err)
	}
	// Delete chat
	if _, err := r.db.NewDelete().Model((*SupportChat)(nil)).Where("id = ?", chatID).Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete support chat: %w", err)
	}
	return nil
}

func (r *Repository) GetSupportMessages(ctx context.Context, chatID uuid.UUID) ([]SupportMessage, error) {
	var msgs []SupportMessage
	err := r.db.NewSelect().Model(&msgs).Where("chat_id = ?", chatID).Order("timestamp ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get support messages: %w", err)
	}
	return msgs, nil
}

func (r *Repository) GetSupportChatByID(ctx context.Context, chatID uuid.UUID) (*SupportChat, error) {
	chat := new(SupportChat)
	if err := r.db.NewSelect().Model(chat).Where("id = ?", chatID).Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get support chat: %w", err)
	}
	return chat, nil
}

func (r *Repository) SaveSupportMessage(ctx context.Context, msg *SupportMessage) error {
	msg.Timestamp = time.Now()
	_, err := r.db.NewInsert().Model(msg).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save support message: %w", err)
	}
	return nil
}

func (r *Repository) UpdateSupportMessage(ctx context.Context, id uint, senderID string, content string) error {
	_, err := r.db.NewUpdate().
		Model((*SupportMessage)(nil)).
		Set("content = ?", content).
		Set("edited = TRUE").
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND sender_id = ?", id, senderID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update support message: %w", err)
	}
	return nil
}

func (r *Repository) DeleteSupportMessage(ctx context.Context, id uint, senderID string) error {
	res, err := r.db.NewDelete().Model((*SupportMessage)(nil)).Where("id = ? AND sender_id = ?", id, senderID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete support message: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("no rows deleted")
	}
	return nil
}

func (r *Repository) GetSupportMessageByID(ctx context.Context, id uint) (*SupportMessage, error) {
	msg := new(SupportMessage)
	if err := r.db.NewSelect().Model(msg).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get support message: %w", err)
	}
	return msg, nil
}

func (r *Repository) SetSupportAttachmentURL(ctx context.Context, id uint, senderID string, url string) error {
	_, err := r.db.NewUpdate().
		Model((*SupportMessage)(nil)).
		Set("attachment_url = ?", url).
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND sender_id = ?", id, senderID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set support attachment: %w", err)
	}
	return nil
}

func (r *Repository) GetSupportAttachmentKeyByID(ctx context.Context, id uint, senderID string) (string, error) {
	var key string
	err := r.db.NewSelect().
		TableExpr("support_messages AS m").
		ColumnExpr("m.attachment_url").
		Where("m.id = ? AND m.sender_id = ?", id, senderID).
		Scan(ctx, &key)
	if err != nil {
		return "", fmt.Errorf("failed to get support attachment key: %w", err)
	}
	return key, nil
}

func (r *Repository) MarkSupportMessagesAsRead(ctx context.Context, chatID uuid.UUID, currentUserID string, currentUserRole string) error {
	if currentUserRole == "admin" || currentUserRole == "main_admin" {
		_, err := r.db.NewUpdate().
			Model((*SupportMessage)(nil)).
			Set("read = true").
			Where("chat_id = ? AND sender_role != ? AND sender_role != ? AND read = false",
				chatID, "admin", "main_admin").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to mark support messages as read: %w", err)
		}
	}
	return nil
}

func (r *Repository) GetUnreadSupportMessagesCount(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*SupportMessage)(nil)).
		Where("read = false AND sender_role != ? AND sender_role != ?", "admin", "main_admin").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread support messages count: %w", err)
	}
	return int64(count), nil
}
