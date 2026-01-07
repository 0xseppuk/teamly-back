package handlers

import (
	"time"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateGameApplicationRequest struct {
	GameID          string    `json:"game_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	MaxPlayers      int       `json:"max_players"`
	MinPlayers      int       `json:"min_players"`
	PrimeTimeStart  time.Time `json:"prime_time_start"`
	PrimeTimeEnd    time.Time `json:"prime_time_end"`
	WithVoiceChat   bool      `json:"with_voice_chat"`
	Platform        string    `json:"platform"`
}

// ApplicationWithUserResponse - заявка с информацией об отклике пользователя
type ApplicationWithUserResponse struct {
	models.GameApplication
	UserHasResponded   bool          `json:"user_has_responded"`
	UserResponseStatus *models.Status `json:"user_response_status,omitempty"`
	UserResponseMessage *string       `json:"user_response_message,omitempty"`
}

// CreateGameApplication создает новую заявку на игру
func CreateGameApplication(c *fiber.Ctx) error {
	// Получаем ID пользователя из контекста (должен быть установлен middleware)
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Парсим тело запроса
	var req CreateGameApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валидация
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}
	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}
	if req.MaxPlayers < req.MinPlayers {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Max players must be greater than or equal to min players",
		})
	}
	if req.MinPlayers < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Min players must be at least 1",
		})
	}

	// Парсим GameID
	parsedGameID, err := uuid.Parse(req.GameID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid game ID",
		})
	}

	// Проверяем существование игры
	var game models.Game
	if err := database.DB.First(&game, parsedGameID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	// Создаем заявку
	application := models.GameApplication{
		UserId:         parsedUserID,
		GameId:         parsedGameID,
		Title:          req.Title,
		Description:    req.Description,
		MaxPlayers:     req.MaxPlayers,
		MinPlayers:     req.MinPlayers,
		PrimeTimeStart: req.PrimeTimeStart,
		PrimeTimeEnd:   req.PrimeTimeEnd,
		WithVoiceChat:  req.WithVoiceChat,
		Platform:       models.Platform(req.Platform),
		IsActive:       true,
		IsFull:         false,
	}

	if err := database.DB.Create(&application).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create application",
		})
	}

	// Загружаем связанные данные
	database.DB.Preload("Game").Preload("User").First(&application, application.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Application created successfully",
		"application": application,
	})
}

// GetUserApplications получает все заявки пользователя
func GetUserApplications(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var applications []models.GameApplication
	result := database.DB.
		Preload("Game").
		Preload("User").
		Where("user_id = ?", parsedUserID).
		Order("created_at DESC").
		Find(&applications)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch applications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"applications": applications,
		"count":        len(applications),
	})
}

// GetApplicationsByUserID получает активные заявки конкретного пользователя (публичный endpoint)
func GetApplicationsByUserID(c *fiber.Ctx) error {
	userIDParam := c.Params("id")
	if userIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	parsedUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	var applications []models.GameApplication
	result := database.DB.
		Preload("Game").
		Preload("User").
		Where("user_id = ? AND is_active = ?", parsedUserID, true).
		Order("created_at DESC").
		Find(&applications)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch applications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"applications": applications,
		"count":        len(applications),
	})
}

// GetAllApplications получает все активные заявки
func GetAllApplications(c *fiber.Ctx) error {
	// Пытаемся получить userID напрямую из токена (без middleware)
	var currentUserID *uuid.UUID
	userID, err := utils.GetUserIDFromContext(c)
	if err == nil {
		currentUserID = &userID
	}

	var applications []models.GameApplication

	query := database.DB.
		Preload("Game").
		Preload("User").
		Where("is_active = ?", true)

	// Фильтры
	if gameID := c.Query("game_id"); gameID != "" {
		parsedGameID, err := uuid.Parse(gameID)
		if err == nil {
			query = query.Where("game_id = ?", parsedGameID)
		}
	}

	if platform := c.Query("platform"); platform != "" {
		query = query.Where("platform = ?", platform)
	}

	if withVoiceChat := c.Query("with_voice_chat"); withVoiceChat != "" {
		if withVoiceChat == "true" {
			query = query.Where("with_voice_chat = ?", true)
		}
	}

	result := query.Order("created_at DESC").Find(&applications)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch applications",
		})
	}

	// Если пользователь авторизован, добавляем информацию об откликах
	var applicationsWithResponse []ApplicationWithUserResponse
	if currentUserID != nil {
		// Получаем все отклики текущего пользователя для этих заявок
		applicationIDs := make([]uuid.UUID, len(applications))
		for i, app := range applications {
			applicationIDs[i] = app.ID
		}

		var responses []models.ApplicationResponse
		database.DB.
			Preload("Conversation.Messages", func(db *gorm.DB) *gorm.DB {
				return db.Order("created_at ASC").Limit(1) // Только первое сообщение
			}).
			Where("user_id = ? AND application_id IN ?", currentUserID, applicationIDs).
			Find(&responses)

		// Создаем map для быстрого поиска
		responseMap := make(map[uuid.UUID]*models.ApplicationResponse)
		for i := range responses {
			responseMap[responses[i].ApplicationID] = &responses[i]
		}

		// Формируем ответ с информацией об откликах
		for _, app := range applications {
			appWithResponse := ApplicationWithUserResponse{
				GameApplication:     app,
				UserHasResponded:    false,
				UserResponseStatus:  nil,
				UserResponseMessage: nil,
			}

			if response, exists := responseMap[app.ID]; exists {
				appWithResponse.UserHasResponded = true
				appWithResponse.UserResponseStatus = &response.Status

				// Добавляем первое сообщение если есть
				if response.Conversation != nil && len(response.Conversation.Messages) > 0 {
					appWithResponse.UserResponseMessage = &response.Conversation.Messages[0].Content
				}
			}

			applicationsWithResponse = append(applicationsWithResponse, appWithResponse)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"applications": applicationsWithResponse,
			"count":        len(applicationsWithResponse),
		})
	}

	// Если не авторизован, возвращаем без информации об откликах
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"applications": applications,
		"count":        len(applications),
	})
}

// GetApplicationByID получает заявку по ID
func GetApplicationByID(c *fiber.Ctx) error {
	appID := c.Params("id")
	if appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Application ID is required",
		})
	}

	parsedID, err := uuid.Parse(appID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid application ID format",
		})
	}

	var application models.GameApplication
	result := database.DB.
		Preload("Game").
		Preload("User").
		First(&application, parsedID)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Application not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"application": application,
	})
}

// UpdateApplication обновляет заявку (только создатель)
func UpdateApplication(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	appID := c.Params("id")
	parsedAppID, err := uuid.Parse(appID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid application ID",
		})
	}

	// Находим заявку
	var application models.GameApplication
	if err := database.DB.First(&application, parsedAppID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Application not found",
		})
	}

	// Проверяем владельца
	if application.UserId != parsedUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to update this application",
		})
	}

	// Парсим тело запроса
	var req CreateGameApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валидация
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}
	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}
	if req.MaxPlayers < req.MinPlayers {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Max players must be greater than or equal to min players",
		})
	}
	if req.MinPlayers < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Min players must be at least 1",
		})
	}

	// Обновляем поля
	application.Title = req.Title
	application.Description = req.Description
	application.MaxPlayers = req.MaxPlayers
	application.MinPlayers = req.MinPlayers
	application.PrimeTimeStart = req.PrimeTimeStart
	application.PrimeTimeEnd = req.PrimeTimeEnd
	application.WithVoiceChat = req.WithVoiceChat
	application.Platform = models.Platform(req.Platform)

	// Сохраняем изменения
	if err := database.DB.Save(&application).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update application",
		})
	}

	// Загружаем связанные данные
	database.DB.Preload("Game").Preload("User").First(&application, application.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Application updated successfully",
		"application": application,
	})
}

// DeleteApplication удаляет заявку (только создатель)
func DeleteApplication(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	appID := c.Params("id")
	parsedAppID, err := uuid.Parse(appID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid application ID",
		})
	}

	var application models.GameApplication
	if err := database.DB.First(&application, parsedAppID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Application not found",
		})
	}

	// Проверяем владельца
	if application.UserId != parsedUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to delete this application",
		})
	}

	if err := database.DB.Delete(&application).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete application",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Application deleted successfully",
	})
}
