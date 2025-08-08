package chat

import (
	"github.com/gofiber/fiber/v2"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

type Handler struct {
	service   *Service
	jwtService *jwt.JWT
}

func NewHandler(app fiber.Router, service *Service, jwtService *jwt.JWT) {
	handler := &Handler{
		service:   service,
		jwtService: jwtService,
	}

	// Группа маршрутов для чата
	chatGroup := app.Group("/chat")
	
	// Middleware для аутентификации
	chatGroup.Use(handler.authMiddleware)

	// Маршруты чата
	chatGroup.Get("/users", handler.GetUsers)
	chatGroup.Get("/messages", handler.GetMessages)
	chatGroup.Post("/messages", handler.SendMessage)
	chatGroup.Get("/unread-count", handler.GetUnreadCount)
}

// authMiddleware проверяет аутентификацию пользователя
func (h *Handler) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authorization header is required",
		})
	}

	// Извлекаем токен из заголовка
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Валидируем токен
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid token",
		})
	}

	// Сохраняем информацию о пользователе в контексте
	c.Locals("user_id", claims.UserID)
	c.Locals("user_role", claims.Role)

	return c.Next()
}

// GetUsers возвращает список пользователей для чата
func (h *Handler) GetUsers(c *fiber.Ctx) error {
	role := c.Query("role")
	if role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role parameter is required",
		})
	}

	users, err := h.service.GetUsers(c.Context(), role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get users: " + err.Error(),
		})
	}

	return c.JSON(UsersResponse{
		Users: users,
	})
}

// GetMessages возвращает сообщения между двумя пользователями
func (h *Handler) GetMessages(c *fiber.Ctx) error {
	targetUserID := c.Query("targetUserId")
	if targetUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "targetUserId parameter is required",
		})
	}

	// Получаем ID текущего пользователя из контекста
	senderID := c.Locals("user_id").(string)
	if senderID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}

	messages, err := h.service.GetMessages(c.Context(), senderID, targetUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get messages: " + err.Error(),
		})
	}

	return c.JSON(MessagesResponse{
		Messages: messages,
	})
}

// SendMessage отправляет новое сообщение
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Получаем ID текущего пользователя из контекста
	senderID := c.Locals("user_id").(string)
	if senderID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}

	message, err := h.service.SendMessage(c.Context(), senderID, req.TargetID, req.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to send message: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    message,
	})
}

// GetUnreadCount возвращает количество непрочитанных сообщений
func (h *Handler) GetUnreadCount(c *fiber.Ctx) error {
	// Получаем ID текущего пользователя из контекста
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}

	count, err := h.service.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get unread count: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"count": count},
	})
}
