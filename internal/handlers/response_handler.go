package handlers

import (
	"time"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateApplicationResponse - создание отклика на заявку
// POST /api/applications/:id/responses
func CreateApplicationResponse(c *fiber.Ctx) error {
	applicationID := c.Params("id")

	// Парсим ID заявки
	appUUID, err := uuid.Parse(applicationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid application ID",
		})
	}

	// Получаем ID текущего пользователя из контекста (из JWT middleware)
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Парсим тело запроса
	var req struct {
		Message string `json:"message" validate:"required,min=10"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валидация: минимум 10 символов
	if len(req.Message) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Message must be at least 10 characters long",
		})
	}

	// Проверяем что заявка существует и активна
	var application models.GameApplication
	if err := database.DB.First(&application, appUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Application not found",
		})
	}

	if !application.IsActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Application is not active",
		})
	}

	// Проверяем что пользователь не автор заявки
	if application.UserId == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot respond to your own application",
		})
	}

	// Проверяем что пользователь еще не откликался
	var existingResponse models.ApplicationResponse
	err = database.DB.Where("application_id = ? AND user_id = ?", appUUID, userID).First(&existingResponse).Error
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You have already responded to this application",
		})
	}

	// Транзакция: создаем Response + Conversation + Message
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Создаем отклик
	response := models.ApplicationResponse{
		ApplicationID: appUUID,
		UserID:        userID,
		Status:        models.StatusPending,
	}
	if err := tx.Create(&response).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create response",
		})
	}

	// 2. Создаем диалог
	now := time.Now()
	conversation := models.Conversation{
		ResponseID:     response.ID,
		Participant1ID: application.UserId, // Автор заявки
		Participant2ID: userID,             // Откликнувшийся
		LastMessageAt:  &now,
		IsArchived:     false,
	}
	if err := tx.Create(&conversation).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create conversation",
		})
	}

	// 3. Создаем первое сообщение (сопроводительное письмо)
	message := models.Message{
		ConversationID: conversation.ID,
		SenderID:       userID,
		Content:        req.Message,
		IsRead:         false,
	}
	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create message",
		})
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// Загружаем связанные данные для ответа
	database.DB.Preload("User").Preload("Application").Preload("Conversation").First(&response, response.ID)

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetApplicationResponses - получить все отклики на заявку (только для автора)
// GET /api/applications/:id/responses
func GetApplicationResponses(c *fiber.Ctx) error {
	applicationID := c.Params("id")

	appUUID, err := uuid.Parse(applicationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid application ID",
		})
	}

	// Получаем ID текущего пользователя
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Проверяем что заявка существует
	var application models.GameApplication
	if err := database.DB.First(&application, appUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Application not found",
		})
	}

	// Проверяем что пользователь - автор заявки
	if application.UserId != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only view responses to your own applications",
		})
	}

	// Получаем отклики с первым сообщением
	var responses []models.ApplicationResponse
	err = database.DB.
		Preload("User").
		Preload("Conversation.Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Limit(1) // Только первое сообщение
		}).
		Where("application_id = ?", appUUID).
		Order("created_at DESC").
		Find(&responses).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch responses",
		})
	}

	return c.JSON(responses)
}

// UpdateResponseStatus - принять/отклонить отклик
// PATCH /api/responses/:id
func UpdateResponseStatus(c *fiber.Ctx) error {
	responseID := c.Params("id")

	respUUID, err := uuid.Parse(responseID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid response ID",
		})
	}

	// Получаем ID текущего пользователя
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Парсим тело запроса
	var req struct {
		Status string `json:"status" validate:"required,oneof=accepted rejected"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Проверяем валидность статуса
	var newStatus models.Status
	switch req.Status {
	case "accepted":
		newStatus = models.StatusAccepted
	case "rejected":
		newStatus = models.StatusRejected
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Must be 'accepted' or 'rejected'",
		})
	}

	// Получаем отклик с заявкой
	var response models.ApplicationResponse
	if err := database.DB.Preload("Application").First(&response, respUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Response not found",
		})
	}

	// Проверяем что пользователь - автор заявки
	if response.Application.UserId != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only the application author can update response status",
		})
	}

	// Обновляем статус
	response.Status = newStatus
	if err := database.DB.Save(&response).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update response status",
		})
	}

	// Если отклонили - архивируем диалог
	if newStatus == models.StatusRejected {
		var conversation models.Conversation
		if err := database.DB.Where("response_id = ?", respUUID).First(&conversation).Error; err == nil {
			conversation.IsArchived = true
			database.DB.Save(&conversation)
		}
	}

	return c.JSON(response)
}

// GetMyResponses - получить мои отклики на чужие заявки
// GET /api/responses/my
func GetMyResponses(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var responses []models.ApplicationResponse
	err = database.DB.
		Preload("Application").
		Preload("Application.Game").
		Preload("Conversation.Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Limit(1)
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&responses).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch responses",
		})
	}

	return c.JSON(responses)
}
