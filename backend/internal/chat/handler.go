package chat

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	websocket "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type ChatHandlerDeps struct {
	ChatService ChatServiceInterface
	JwtService  JWTServiceInterface
	Logger      *middleware.Logger
}

type ChatHandler struct {
	fiber.Router
	deps *ChatHandlerDeps
}

func NewChatHandler(router fiber.Router, deps *ChatHandlerDeps) {
	handler := &ChatHandler{
		Router: router,
		deps:   deps,
	}

	jwtConcrete, ok := handler.deps.JwtService.(*jwt.JWT)
	if !ok {
		panic("JwtService must be *jwt.JWT for middleware")
	}

	// ================================================================
	// PUBLIC SUPPORT ROUTES: /api/public/support/*
	// Access: public (no authentication required)
	// ================================================================
	router.Post("/public/support/start", handler.PublicStartSupportChat)                           // POST /api/public/support/start
	router.Get("/public/support/messages", handler.GetSupportMessages)                             // GET /api/public/support/messages
	router.Post("/public/support/messages", handler.SendSupportMessage)                            // POST /api/public/support/messages
	router.Post("/public/support/messages/:id/attachments", handler.UploadPublicSupportAttachment) // POST /api/public/support/messages/:id/attachments
	router.Get("/public/support/messages/:id/attachments", handler.StreamSupportAttachment)        // GET /api/public/support/messages/:id/attachments

	// ================================================================
	// WEBSOCKET ROUTES: /api/ws/*
	// Access: public (websocket connection)
	// ================================================================
	hub := NewHub()
	handler.deps.ChatService.AttachHub(hub)
	router.Get("/ws/chat", websocket.New(handler.Socket)) // GET /api/ws/chat

	// ================================================================
	// AUTHENTICATED CHAT ROUTES: /api/chat/*
	// Access: authenticated users (admin, partner, or guest)
	// ================================================================
	chatGroup := router.Group("/chat")
	chatGroup.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger))

	chatGroup.Get("/users", handler.GetUsers)                             // GET /api/chat/users
	chatGroup.Get("/messages", handler.GetMessages)                       // GET /api/chat/messages
	chatGroup.Post("/messages", handler.SendMessage)                      // POST /api/chat/messages
	chatGroup.Get("/unread-count", handler.GetUnreadCount)                // GET /api/chat/unread-count
	chatGroup.Get("/unread-by-sender", handler.GetUnreadBySender)         // GET /api/chat/unread-by-sender
	chatGroup.Get("/support-unread-count", handler.GetSupportUnreadCount) // GET /api/chat/support-unread-count
	chatGroup.Patch("/messages/:id", handler.UpdateMessage)               // PATCH /api/chat/messages/:id
	chatGroup.Delete("/messages/:id", handler.DeleteMessage)              // DELETE /api/chat/messages/:id
	chatGroup.Post("/messages/:id/attachments", handler.UploadAttachment) // POST /api/chat/messages/:id/attachments
	chatGroup.Patch("/partners/:id/block", handler.BlockPartner)          // PATCH /api/chat/partners/:id/block
	chatGroup.Patch("/partners/:id/unblock", handler.UnblockPartner)      // PATCH /api/chat/partners/:id/unblock

	// ================================================================
	// AUTHENTICATED ATTACHMENT ROUTES: /api/public/attachments/*
	// Access: authenticated users (admin, partner, or guest)
	// ================================================================
	attachments := router.Group("/public/attachments")
	attachments.Get(":id", handler.StreamAttachmentByMessageID) // GET /api/public/attachments/:id

	// ================================================================
	// AUTHENTICATED SUPPORT ROUTES: /api/support/*
	// Access: authenticated users (admin, partner, or guest)
	// ================================================================
	support := router.Group("/support")
	support.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger))
	support.Get("/messages", handler.GetSupportMessages)                          // GET /api/support/messages
	support.Post("/messages", handler.SendSupportMessage)                         // POST /api/support/messages
	support.Post("/messages/:id/attachments", handler.UploadSupportAttachment)    // POST /api/support/messages/:id/attachments
	support.Get("/messages/:id/attachments", handler.StreamAuthSupportAttachment) // GET /api/support/messages/:id/attachments
	support.Patch("/messages/:id", handler.UpdateSupportMessageHandler)           // PATCH /api/support/messages/:id
	support.Delete("/messages/:id", handler.DeleteSupportMessageHandler)          // DELETE /api/support/messages/:id

	// ================================================================
	// ADMIN SUPPORT ROUTES: /api/admin/support/*
	// Access: admin and main_admin roles only
	// ================================================================
	adminSupport := router.Group("/admin/support")
	adminSupport.Use(middleware.JWTMiddleware(jwtConcrete, deps.Logger), middleware.AdminOrMainAdmin())
	adminSupport.Get("/chats", handler.AdminListSupportChats)         // GET /api/admin/support/chats
	adminSupport.Delete("/chats/:id", handler.AdminDeleteSupportChat) // DELETE /api/admin/support/chats/:id
}

// @Summary      Start a support chat
// @Description  Creates a new support chat and returns a guest token for access
// @Tags         support
// @Produce      json
// @Param        title  query     string  true  "Title of the support chat"
// @Success      200    {object} map[string]any "Chat details including guest token"
// @Failure      500    {object} map[string]any "Internal server error"
// @Router       /public/support/start [post]
func (handler *ChatHandler) PublicStartSupportChat(c *fiber.Ctx) error {
	title := strings.TrimSpace(c.Query("title"))
	guestID := uuid.New()
	chat, err := handler.deps.ChatService.StartSupportChat(c.Context(), title, guestID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to start support chat")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to start support chat",
		})
	}
	access, err := handler.deps.JwtService.CreateAccessToken(guestID, "guest", "user")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to create access token")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to create access token",
		})
	}
	return c.JSON(map[string]any{"chat_id": chat.ID, "title": chat.Title, "guest_id": guestID, "access_token": access})
}

// @Summary      Get support chat messages
// @Description  Retrieves messages for a specific support chat by chat ID
// @Tags         support
// @Produce      json
// @Param        chat_id  query     string  true  "ID of the chat"
// @Success      200      {object} map[string]any "List of support messages"
// @Failure      400      {object} map[string]any "Invalid chat ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      404      {object} map[string]any "Chat not found"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /public/support/messages [get]
func (handler *ChatHandler) GetSupportMessages(c *fiber.Ctx) error {
	chatIDStr := c.Query("chat_id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid chat_id")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid chat_id",
		})
	}

	var role string
	var userID string

	if c.Locals("user_role") != nil && c.Locals("user_id") != nil {
		role = c.Locals("user_role").(string)
		userID = fmt.Sprintf("%v", c.Locals("user_id"))
	} else {
		authHeader := c.Get("Authorization")
		tokenCandidate := authHeader
		if tokenCandidate == "" {
			tokenCandidate = c.Query("token")
		}
		if tokenCandidate == "" {
			handler.deps.Logger.FromContext(c).Error().Msg("Authorization header is required")
			return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
				"error": "Authorization header is required",
			})
		}

		token := tokenCandidate
		if len(tokenCandidate) > 7 && tokenCandidate[:7] == "Bearer " {
			token = tokenCandidate[7:]
		}

		claims, err := handler.deps.JwtService.ValidateAccessToken(token)
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
				"error": "Invalid token",
			})
		}

		role = claims.Role
		userID = claims.UserID.String()
		handler.deps.Logger.FromContext(c).Info().Str("role", role).Str("userID", userID).Msg("GetSupportMessages: Using manual JWT parsing")
	}

	if role != "admin" && role != "main_admin" {
		chat, err := handler.deps.ChatService.GetChatRepository().GetSupportChatByID(c.Context(), chatID)
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Chat not found")
			return c.Status(fiber.StatusNotFound).JSON(map[string]any{
				"error": "Chat not found",
			})
		}
		if chat.GuestID.String() != userID {
			handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("forbidden")).Msg("Access denied")
			return c.Status(fiber.StatusForbidden).JSON(map[string]any{
				"error": "Access denied",
			})
		}
	}
	msgs, err := handler.deps.ChatService.GetSupportMessages(c.Context(), chatID, userID, role)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get messages")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to get messages",
		})
	}
	return c.JSON(map[string]any{"messages": msgs})
}

// @Summary      Send a message in support chat
// @Description  Sends a new message in a specific support chat
// @Tags         support
// @Produce      json
// @Param        chat_id  query     string  true  "ID of the chat"
// @Param        content  query     string  true  "Content of the message"
// @Success      200      {object} map[string]any "Sent message details"
// @Failure      400      {object} map[string]any "Invalid request body or chat ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      403      {object} map[string]any "Access denied"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /public/support/messages [post]
func (handler *ChatHandler) SendSupportMessage(c *fiber.Ctx) error {
	var body struct {
		ChatID  string `json:"chat_id"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid body")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid body",
		})
	}
	chatID, err := uuid.Parse(body.ChatID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid chat_id")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid chat_id",
		})
	}

	authHeader := c.Get("Authorization")
	tokenCandidate := authHeader
	if tokenCandidate == "" {
		tokenCandidate = c.Query("token")
	}
	if tokenCandidate == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Authorization header is required")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Authorization header is required",
		})
	}

	token := tokenCandidate
	if len(tokenCandidate) > 7 && tokenCandidate[:7] == "Bearer " {
		token = tokenCandidate[7:]
	}

	claims, err := handler.deps.JwtService.ValidateAccessToken(token)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid token")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Invalid token",
		})
	}

	role := claims.Role
	sender := claims.UserID.String()
	if role != "admin" && role != "main_admin" {
		chat, err := handler.deps.ChatService.GetChatRepository().GetSupportChatByID(c.Context(), chatID)
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Chat not found")
			return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
				"error": "Chat not found",
			})
		}
		if chat.GuestID.String() != sender {
			handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("forbidden")).Msg("Access denied")
			return c.Status(fiber.StatusForbidden).JSON(map[string]any{
				"error": "Access denied",
			})
		}
	}
	msg, err := handler.deps.ChatService.SendSupportMessage(c.Context(), chatID, sender, role, body.Content)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to send message")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to send message",
		})
	}
	return c.JSON(map[string]any{"message": msg})
}

// @Summary      Update a support message
// @Description  Updates the content of a specific support message
// @Tags         support
// @Produce      json
// @Param        id      path      int     true  "ID of the message"
// @Param        content query     string  true  "New content for the message"
// @Success      200      {object} map[string]any "Update success confirmation"
// @Failure      400      {object} map[string]any "Invalid request body or ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /support/messages/{id} [patch]
func (handler *ChatHandler) UpdateSupportMessageHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid id")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid id",
		})
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid body")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid body",
		})
	}
	sender := fmt.Sprintf("%v", c.Locals("user_id"))
	if err := handler.deps.ChatService.UpdateSupportMessage(c.Context(), uint(id), sender, body.Content); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update message")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to update message",
		})
	}
	return c.JSON(map[string]any{"success": true})
}

// @Summary      Delete a support message
// @Description  Deletes a specific support message by ID
// @Tags         support
// @Produce      json
// @Param        id  path      int  true  "ID of the message"
// @Success      200      {object} map[string]any "Delete success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /support/messages/{id} [delete]
func (handler *ChatHandler) DeleteSupportMessageHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid id")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid id",
		})
	}
	sender := fmt.Sprintf("%v", c.Locals("user_id"))
	if err := handler.deps.ChatService.DeleteSupportMessage(c.Context(), uint(id), sender); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete message")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to delete message",
		})
	}
	return c.JSON(map[string]any{"success": true})
}

// @Summary      Admin delete support chat
// @Description  Deletes a support chat by ID (admin only)
// @Tags         admin, support
// @Produce      json
// @Param        id  path      int  true  "ID of the chat"
// @Success      200      {object} map[string]any "Delete success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      403      {object} map[string]any "Access denied"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /admin/support/chats/{id} [delete]
func (handler *ChatHandler) AdminDeleteSupportChat(c *fiber.Ctx) error {
	if !handler.checkAdmin(c) {
		handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("forbidden")).Msg("Access denied")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied",
		})
	}
	idStr := c.Params("id")
	chatID, err := uuid.Parse(idStr)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid id")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid id",
		})
	}
	if err := handler.deps.ChatService.DeleteSupportChat(c.Context(), chatID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete support chat")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to delete support chat",
		})
	}
	return c.JSON(map[string]any{"success": true})
}

func (handler *ChatHandler) checkAdmin(c *fiber.Ctx) bool {
	role, _ := c.Locals("user_role").(string)

	// Also try to get from jwt_claims
	if claims, ok := c.Locals("jwt_claims").(*jwt.Claims); ok && claims != nil {
		return claims.Role == "admin" || claims.Role == "main_admin"
	}

	return role == "admin" || role == "main_admin"
}

// @Summary      Admin block partner
// @Description  Blocks a partner by ID (admin only)
// @Tags         admin, support
// @Produce      json
// @Param        id  path      int  true  "ID of the partner"
// @Success      200      {object} map[string]any "Block success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      403      {object} map[string]any "Access denied"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /admin/support/partners/{id}/block [patch]
func (handler *ChatHandler) BlockPartner(c *fiber.Ctx) error {
	if !handler.checkAdmin(c) {
		handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("Forbidden")).Msg("Access denied")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied",
		})
	}
	partnerID := c.Params("id")
	if err := handler.deps.ChatService.AdminBlockPartner(c.Context(), partnerID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to block partner")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to block partner",
		})
	}
	return c.JSON(map[string]any{"success": true})
}

// @Summary      Admin unblock partner
// @Description  Unblocks a partner by ID (admin only)
// @Tags         admin, support
// @Produce      json
// @Param        id  path      int  true  "ID of the partner"
// @Success      200      {object} map[string]any "Unblock success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      403      {object} map[string]any "Access denied"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /admin/support/partners/{id}/unblock [patch]
func (handler *ChatHandler) UnblockPartner(c *fiber.Ctx) error {
	if !handler.checkAdmin(c) {
		handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("Forbidden")).Msg("Access denied")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied",
		})
	}
	partnerID := c.Params("id")
	if err := handler.deps.ChatService.AdminUnblockPartner(c.Context(), partnerID); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to unblock partner")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to unblock partner",
		})
	}
	return c.JSON(map[string]any{"success": true})
}

// @Summary      Authenticate user
// @Description  Middleware to authenticate user using JWT
// @Tags         auth
// @Produce      json

// @Summary      Get users by role
// @Description  Retrieves a list of users filtered by role
// @Tags         users
// @Produce      json
// @Param        role   query     string  true  "Role of the users"
// @Param        search query     string  false "Search term"
// @Success      200    {object} UsersResponse "List of users"
// @Failure      400    {object} map[string]any "Invalid role parameter"
// @Failure      500    {object} map[string]any "Internal server error"
// @Router       /chat/users [get]
func (handler *ChatHandler) GetUsers(c *fiber.Ctx) error {
	role := c.Query("role")
	if role == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Role parameter is required")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Role parameter is required",
		})
	}

	search := c.Query("search")
	users, err := handler.deps.ChatService.GetUsers(c.Context(), role, search)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get users")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to get users",
		})
	}

	return c.JSON(UsersResponse{
		Users: users,
	})
}

// @Summary      Get messages between users
// @Description  Retrieves messages between the authenticated user and a target user
// @Tags         messages
// @Produce      json
// @Param        targetUserId  query     string  true  "ID of the target user"
// @Success      200      {object} MessagesResponse "List of messages"
// @Failure      400      {object} map[string]any "Invalid targetUserId parameter"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/messages [get]
func (handler *ChatHandler) GetMessages(c *fiber.Ctx) error {
	targetUserID := c.Query("targetUserId")
	if targetUserID == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("TargetUserId parameter is required")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "TargetUserId parameter is required",
		})
	}

	uidVal := c.Locals("user_id")
	var senderID string
	switch v := uidVal.(type) {
	case string:
		senderID = v
	default:
		senderID = fmt.Sprintf("%v", v)
	}
	if senderID == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "User not authenticated",
		})
	}

	messages, err := handler.deps.ChatService.GetMessages(c.Context(), senderID, targetUserID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get messages")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to get messages",
		})
	}

	messageResponses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = MessageResponse{
			ID:            msg.ID,
			SenderID:      msg.SenderID,
			TargetID:      msg.TargetID,
			Content:       msg.Content,
			Timestamp:     msg.Timestamp,
			Read:          msg.Read,
			AttachmentURL: msg.AttachmentURL,
			Edited:        msg.Edited,
			UpdatedAt:     msg.UpdatedAt,
		}
	}

	return c.JSON(MessagesResponse{
		Messages: messageResponses,
	})
}

// @Summary      Send a message to a user
// @Description  Sends a new message to a specific user
// @Tags         messages
// @Produce      json
// @Param        targetId  query     string  true  "ID of the target user"
// @Param        content   query     string  true  "Content of the message"
// @Success      200      {object} map[string]any "Sent message details"
// @Failure      400      {object} map[string]any "Invalid request body or target ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/messages [post]
func (handler *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid request body",
		})
	}

	uidVal := c.Locals("user_id")
	var senderID string
	switch v := uidVal.(type) {
	case string:
		senderID = v
	default:
		senderID = fmt.Sprintf("%v", v)
	}
	if senderID == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "User not authenticated",
		})
	}

	message, err := handler.deps.ChatService.SendMessage(c.Context(), senderID, req.TargetID, req.Content)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to send message")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to send message",
		})
	}

	return c.JSON(map[string]any{
		"success": true,
		"data":    message,
	})
}

// @Summary      Get unread message count
// @Description  Retrieves the count of unread messages for the authenticated user
// @Tags         messages
// @Produce      json
// @Success      200      {object} map[string]any "Unread count details"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/unread-count [get]
func (handler *ChatHandler) GetUnreadCount(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	var userID string
	switch v := uidVal.(type) {
	case string:
		userID = v
	default:
		userID = fmt.Sprintf("%v", v)
	}
	if userID == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "User not authenticated",
		})
	}

	count, err := handler.deps.ChatService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get unread count")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to get unread count",
		})
	}

	return c.JSON(map[string]any{
		"success": true,
		"data":    map[string]any{"count": count},
	})
}

// @Summary      Get unread count by sender
// @Description  Retrieves unread message count grouped by sender for the authenticated user
// @Tags         messages
// @Produce      json
// @Success      200      {object} map[string]any "Unread count by sender"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/unread-by-sender [get]
func (handler *ChatHandler) GetUnreadBySender(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	userID := fmt.Sprintf("%v", uidVal)
	if userID == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "User not authenticated",
		})
	}
	rows, err := handler.deps.ChatService.GetUnreadBySender(c.Context(), userID)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get unread count")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to get unread count",
		})
	}
	return c.JSON(map[string]any{"success": true, "data": rows})
}

// @Summary      Get support unread count
// @Description  Retrieves the count of unread support messages for admins
// @Tags         messages
// @Produce      json
// @Success      200      {object} map[string]any "Support unread count details"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      403      {object} map[string]any "Forbidden (admins only)"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/support-unread-count [get]
func (handler *ChatHandler) GetSupportUnreadCount(c *fiber.Ctx) error {
	roleVal := c.Locals("user_role")
	role := fmt.Sprintf("%v", roleVal)

	if role != "admin" && role != "main_admin" {
		handler.deps.Logger.FromContext(c).Error().Msg("Access denied: support unread count is for admins only")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied: support unread count is for admins only",
		})
	}

	count, err := handler.deps.ChatService.GetChatRepository().GetUnreadSupportMessagesCount(c.Context())
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to get support unread count")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to get support unread count",
		})
	}

	return c.JSON(map[string]any{
		"success": true,
		"data":    map[string]any{"count": count},
	})
}

// @Summary      Update a message
// @Description  Updates the content of a message by ID
// @Tags         messages
// @Produce      json
// @Param        id      path      int     true  "ID of the message"
// @Param        content query     string  true  "New content for the message"
// @Success      200      {object} map[string]any "Update success confirmation"
// @Failure      400      {object} map[string]any "Invalid request body or ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/messages/{id} [patch]
func (handler *ChatHandler) UpdateMessage(c *fiber.Ctx) error {
	idParam := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil || body.Content == "" {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid request body",
		})
	}
	uidVal := c.Locals("user_id")
	senderID := fmt.Sprintf("%v", uidVal)
	if err := handler.deps.ChatService.GetChatRepository().UpdateMessage(c.Context(), id, senderID, body.Content); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to update message")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to update message",
		})
	}
	return c.JSON(map[string]any{"success": true})
}

// @Summary      Delete a message
// @Description  Deletes a message by ID
// @Tags         messages
// @Produce      json
// @Param        id  path      int  true  "ID of the message"
// @Success      200      {object} map[string]any "Delete success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/messages/{id} [delete]
func (handler *ChatHandler) DeleteMessage(c *fiber.Ctx) error {
	idParam := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}
	uidVal := c.Locals("user_id")
	senderID := fmt.Sprintf("%v", uidVal)
	msg, err := handler.deps.ChatService.GetChatRepository().GetMessageByID(c.Context(), id)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Message not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Message not found",
		})
	}
	if msg.SenderID != senderID {
		handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("Forbidden")).Msg("Access denied")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied",
		})
	}
	key := msg.AttachmentURL
	if err := handler.deps.ChatService.GetChatRepository().DeleteMessageByID(c.Context(), id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to delete message")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to delete message",
		})
	}
	if key != "" && handler.deps.ChatService.GetS3Client() != nil {
		_ = handler.deps.ChatService.GetS3Client().DeleteChatData(c.Context(), key)
	}
	return c.JSON(map[string]any{"success": true})
}

// @Summary      Upload an attachment to a message
// @Description  Uploads an attachment file to a specific message
// @Tags         messages
// @Produce      json
// @Param        id  path      int  true  "ID of the message"
// @Param        file  formData  file  true  "Attachment file"
// @Success      200      {object} map[string]any "Upload success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID or file"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/messages/{id}/attachments [post]
func (handler *ChatHandler) UploadAttachment(c *fiber.Ctx) error {
	idParam := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}
	msg, err := handler.deps.ChatService.GetChatRepository().GetMessageByID(c.Context(), id)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Message not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Message not found",
		})
	}
	uidVal := c.Locals("user_id")
	senderID := fmt.Sprintf("%v", uidVal)
	if msg.SenderID != senderID {
		handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("Forbidden")).Msg("Access denied")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied",
		})
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("File is required")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "File is required",
		})
	}
	f, err := fileHeader.Open()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to open file")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to open file",
		})
	}
	defer f.Close()
	url, upErr := handler.deps.ChatService.UploadAttachment(c.Context(), id, senderID, f, fileHeader.Size, fileHeader.Header.Get("Content-Type"), fileHeader.Filename)
	if upErr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(upErr).Msg("Failed to upload attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to upload attachment",
		})
	}
	return c.JSON(map[string]any{"success": true, "attachment_url": url})
}

// @Summary      Stream an attachment by message ID
// @Description  Streams an attachment file for a specific message by ID
// @Tags         messages
// @Produce      octet-stream
// @Param        id  path      int  true  "ID of the message"
// @Success      200      {file}  "Attachment file"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      404      {object} map[string]any "Attachment not found"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /chat/messages/{id}/attachments [get]
func (handler *ChatHandler) StreamAttachmentByMessageID(c *fiber.Ctx) error {
	var id uint
	if _, err := fmt.Sscanf(c.Params("id"), "%d", &id); err != nil || id == 0 {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}

	// Authenticate using token query parameter
	tokenCandidate := c.Query("token")
	if tokenCandidate == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Authentication required - token parameter missing")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Authentication required",
		})
	}

	_, err := handler.deps.JwtService.ValidateAccessToken(tokenCandidate)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid token")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Invalid token",
		})
	}

	// Try to get regular message first
	msg, err := handler.deps.ChatService.GetChatRepository().GetMessageByID(c.Context(), id)
	if err != nil || msg.AttachmentURL == "" {
		// If not found, try to get support message
		supportMsg, supportErr := handler.deps.ChatService.GetChatRepository().GetSupportMessageByID(c.Context(), id)
		if supportErr != nil || supportMsg.AttachmentURL == "" {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Attachment not found")
			return c.Status(fiber.StatusNotFound).JSON(map[string]any{
				"error": "Attachment not found",
			})
		}
		// Use support message
		rc, ct, derr := handler.deps.ChatService.DownloadSupportAttachment(c.Context(), supportMsg.AttachmentURL)
		if derr != nil {
			handler.deps.Logger.FromContext(c).Error().Err(derr).Msg("Attachment not found")
			return c.Status(fiber.StatusNotFound).JSON(map[string]any{
				"error": "Attachment not found",
			})
		}
		defer rc.Close()
		data, readErr := io.ReadAll(rc)
		if readErr != nil {
			handler.deps.Logger.FromContext(c).Error().Err(readErr).Msg("Failed to read attachment")
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
				"error": "Failed to read attachment",
			})
		}
		if ct == "" {
			ct = "application/octet-stream"
		}
		c.Set("Content-Type", ct)
		c.Set("Content-Length", strconv.Itoa(len(data)))
		c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", sanitizeFilenameFromKey(supportMsg.AttachmentURL)))
		return c.Send(data)
	}

	// Use regular message
	rc, ct, derr := handler.deps.ChatService.DownloadAttachment(c.Context(), msg.AttachmentURL)
	if derr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(derr).Msg("Attachment not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Attachment not found",
		})
	}
	defer rc.Close()
	data, readErr := io.ReadAll(rc)
	if readErr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(readErr).Msg("Failed to read attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to read attachment",
		})
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Set("Content-Type", ct)
	c.Set("Content-Length", strconv.Itoa(len(data)))
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", sanitizeFilenameFromKey(msg.AttachmentURL)))
	return c.Send(data)
}

func sanitizeFilenameFromKey(key string) string {
	parts := strings.Split(key, "/")
	last := "attachment"
	if len(parts) > 0 {
		last = parts[len(parts)-1]
	}
	idx := strings.Index(last, "_")
	if idx >= 0 && idx+1 < len(last) {
		last = last[idx+1:]
	}
	last = strings.ReplaceAll(last, "\"", "")
	last = strings.ReplaceAll(last, "\n", "")
	last = strings.ReplaceAll(last, "\r", "")
	if last == "" {
		last = "attachment"
	}
	return last
}

// @Summary      List support chats
// @Description  Lists all support chats with details (admin only)
// @Tags         admin, support
// @Produce      json
// @Success      200      {object} map[string]any "List of support chats"
// @Failure      403      {object} map[string]any "Access denied"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /admin/support/chats [get]
func (handler *ChatHandler) AdminListSupportChats(c *fiber.Ctx) error {
	result := handler.checkAdmin(c)
	if !result {
		handler.deps.Logger.FromContext(c).Error().Err(fmt.Errorf("Forbidden")).Msg("Access denied")
		return c.Status(fiber.StatusForbidden).JSON(map[string]any{
			"error": "Access denied",
		})
	}
	chats, err := handler.deps.ChatService.ListSupportChats(c.Context())
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to list support chats")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to list support chats",
		})
	}
	return c.JSON(map[string]any{"chats": chats})
}

// @Summary      Upload an attachment to a support message (authenticated users)
// @Description  Uploads an attachment file to a specific support message (requires JWT authentication)
// @Tags         support
// @Produce      json
// @Param        id  path      int  true  "ID of the support message"
// @Param        file  formData  file  true  "Attachment file"
// @Success      200      {object} map[string]any "Upload success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID or file"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /support/messages/{id}/attachments [post]
func (handler *ChatHandler) UploadSupportAttachment(c *fiber.Ctx) error {
	idParam := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}

	// Simplified: only works with JWT middleware
	uidVal := c.Locals("user_id")
	if uidVal == nil {
		handler.deps.Logger.FromContext(c).Error().Msg("User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Authentication required",
		})
	}
	sender := fmt.Sprintf("%v", uidVal)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("File is required")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "File is required",
		})
	}
	handler.deps.Logger.FromContext(c).Info().
		Str("filename", fileHeader.Filename).
		Int64("size", fileHeader.Size).
		Str("content_type", fileHeader.Header.Get("Content-Type")).
		Uint("message_id", id).
		Str("sender_id", sender).
		Msg("Support attachment upload started (authenticated)")
	f, err := fileHeader.Open()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to open file")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to open file",
		})
	}
	defer f.Close()
	attachmentURL, upErr := handler.deps.ChatService.UploadSupportAttachment(c.Context(), id, sender, f, fileHeader.Size, fileHeader.Header.Get("Content-Type"), fileHeader.Filename)
	if upErr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(upErr).Msg("Failed to upload support attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to upload support attachment",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Uint("message_id", id).Str("sender_id", sender).Str("attachment_url", attachmentURL).Msg("Support attachment uploaded successfully")
	return c.JSON(map[string]any{"success": true, "attachment_url": attachmentURL})
}

// @Summary      Upload an attachment to a support message (public users)
// @Description  Uploads an attachment file to a specific support message (uses access token)
// @Tags         support
// @Produce      json
// @Param        id  path      int  true  "ID of the support message"
// @Param        file  formData  file  true  "Attachment file"
// @Success      200      {object} map[string]any "Upload success confirmation"
// @Failure      400      {object} map[string]any "Invalid ID or file"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /public/support/messages/{id}/attachments [post]
func (handler *ChatHandler) UploadPublicSupportAttachment(c *fiber.Ctx) error {
	idParam := c.Params("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}

	// Public access: only access token validation
	authHeader := c.Get("Authorization")
	tokenCandidate := authHeader
	if tokenCandidate == "" {
		tokenCandidate = c.Query("token")
	}
	if tokenCandidate == "" {
		handler.deps.Logger.FromContext(c).Error().Msg("Authorization header is required")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Authorization header is required",
		})
	}
	token := tokenCandidate
	if len(tokenCandidate) > 7 && tokenCandidate[:7] == "Bearer " {
		token = tokenCandidate[7:]
	}
	claims, err := handler.deps.JwtService.ValidateAccessToken(token)
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid token")
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
			"error": "Invalid token",
		})
	}
	sender := claims.UserID.String()

	fileHeader, err := c.FormFile("file")
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("File is required")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "File is required",
		})
	}
	handler.deps.Logger.FromContext(c).Info().
		Str("filename", fileHeader.Filename).
		Int64("size", fileHeader.Size).
		Str("content_type", fileHeader.Header.Get("Content-Type")).
		Uint("message_id", id).
		Str("sender_id", sender).
		Msg("Support attachment upload started (public)")
	f, err := fileHeader.Open()
	if err != nil {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Failed to open file")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Failed to open file",
		})
	}
	defer f.Close()
	attachmentURL, upErr := handler.deps.ChatService.UploadSupportAttachment(c.Context(), id, sender, f, fileHeader.Size, fileHeader.Header.Get("Content-Type"), fileHeader.Filename)
	if upErr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(upErr).Msg("Failed to upload support attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to upload support attachment",
		})
	}
	handler.deps.Logger.FromContext(c).Info().Uint("message_id", id).Str("sender_id", sender).Str("attachment_url", attachmentURL).Msg("Public support attachment uploaded successfully")
	return c.JSON(map[string]any{"success": true, "attachment_url": attachmentURL})
}

// @Summary      Stream an attachment of a support message (authenticated users)
// @Description  Streams an attachment file for a specific support message by ID (requires JWT authentication)
// @Tags         support
// @Produce      octet-stream
// @Param        id  path      int  true  "ID of the support message"
// @Success      200      {file}  "Attachment file"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      401      {object} map[string]any "Unauthorized"
// @Failure      404      {object} map[string]any "Attachment not found"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /support/messages/{id}/attachments [get]
func (handler *ChatHandler) StreamAuthSupportAttachment(c *fiber.Ctx) error {
	var id uint
	if _, err := fmt.Sscanf(c.Params("id"), "%d", &id); err != nil || id == 0 {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}

	// Support both JWT middleware and token query parameter
	if c.Locals("user_id") == nil {
		// No JWT middleware auth - fallback to token query parameter (for direct browser links)
		tokenCandidate := c.Query("token")
		if tokenCandidate == "" {
			handler.deps.Logger.FromContext(c).Error().Msg("Authentication required")
			return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
				"error": "Authentication required",
			})
		}
		_, err := handler.deps.JwtService.ValidateAccessToken(tokenCandidate)
		if err != nil {
			handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
				"error": "Invalid token",
			})
		}
	}
	// If we reach here, user is authenticated either via JWT middleware or token query parameter

	msg, err := handler.deps.ChatService.GetChatRepository().GetSupportMessageByID(c.Context(), id)
	if err != nil || msg.AttachmentURL == "" {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Attachment not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Attachment not found",
		})
	}
	rc, ct, derr := handler.deps.ChatService.DownloadSupportAttachment(c.Context(), msg.AttachmentURL)
	if derr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(derr).Msg("Attachment not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Attachment not found",
		})
	}
	defer rc.Close()
	data, readErr := io.ReadAll(rc)
	if readErr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(readErr).Msg("Failed to read attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to read attachment",
		})
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Set("Content-Type", ct)
	c.Set("Content-Length", strconv.Itoa(len(data)))
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", sanitizeFilenameFromKey(msg.AttachmentURL)))
	return c.Send(data)
}

// @Summary      Stream an attachment of a support message (public users)
// @Description  Streams an attachment file for a specific support message by ID (uses access token)
// @Tags         support
// @Produce      octet-stream
// @Param        id  path      int  true  "ID of the support message"
// @Success      200      {file}  "Attachment file"
// @Failure      400      {object} map[string]any "Invalid ID"
// @Failure      404      {object} map[string]any "Attachment not found"
// @Failure      500      {object} map[string]any "Internal server error"
// @Router       /public/support/messages/{id}/attachments [get]
func (handler *ChatHandler) StreamSupportAttachment(c *fiber.Ctx) error {
	var id uint
	if _, err := fmt.Sscanf(c.Params("id"), "%d", &id); err != nil || id == 0 {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Invalid ID")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"error": "Invalid ID",
		})
	}
	msg, err := handler.deps.ChatService.GetChatRepository().GetSupportMessageByID(c.Context(), id)
	if err != nil || msg.AttachmentURL == "" {
		handler.deps.Logger.FromContext(c).Error().Err(err).Msg("Attachment not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Attachment not found",
		})
	}
	rc, ct, derr := handler.deps.ChatService.DownloadSupportAttachment(c.Context(), msg.AttachmentURL)
	if derr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(derr).Msg("Attachment not found")
		return c.Status(fiber.StatusNotFound).JSON(map[string]any{
			"error": "Attachment not found",
		})
	}
	defer rc.Close()
	data, readErr := io.ReadAll(rc)
	if readErr != nil {
		handler.deps.Logger.FromContext(c).Error().Err(readErr).Msg("Failed to read attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]any{
			"error": "Failed to read attachment",
		})
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Set("Content-Type", ct)
	c.Set("Content-Length", strconv.Itoa(len(data)))
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", sanitizeFilenameFromKey(msg.AttachmentURL)))
	return c.Send(data)
}
