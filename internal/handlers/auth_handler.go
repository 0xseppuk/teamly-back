package handlers

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// setAuthCookie устанавливает HTTP-only cookie с токеном авторизации
func setAuthCookie(c *fiber.Ctx, token string) {
	isProduction := os.Getenv("GO_ENV") == "production"

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 24 часа
		HTTPOnly: true,
		Secure:   isProduction, // true только в production с HTTPS
		SameSite: "Lax",
	})
}

func RegisterUser(c *fiber.Ctx) error {
	var req models.AuthRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("[Register] Body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" || req.Nickname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, nickname and password are required",
		})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("[Register] Hash password error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Nickname:     req.Nickname,
	}

	res := database.DB.Create(&user)
	if res.Error != nil {
		log.Printf("[Register] DB create error: %v", res.Error)
		// Check for duplicate email/nickname
		if strings.Contains(res.Error.Error(), "idx_users_email") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already registered",
			})
		}
		if strings.Contains(res.Error.Error(), "idx_users_nickname") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Nickname already taken",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	log.Printf("[Register] User created: %s", user.ID)

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		log.Printf("[Register] Token generation error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	log.Printf("[Register] Success for user: %s", user.Email)

	// Устанавливаем HTTP-only cookie
	setAuthCookie(c, token)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    user,
	})
}

func LoginUser(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	user := models.User{}
	database.DB.Where("email = ?", req.Email).First(&user)

	if user.ID == uuid.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Email not found",
		})
	}

	if !utils.ComparePassword(user.PasswordHash, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect password",
		})
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Устанавливаем HTTP-only cookie
	setAuthCookie(c, token)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"user":    user,
	})
}

func GetMe(c *fiber.Ctx) error {
	// Получаем user ID из cookie или Authorization header
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"details": err.Error(),
		})
	}

	user := models.User{}
	result := database.DB.Preload("Country").Where("id = ?", userID).First(&user)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "User not found",
			"details": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user,
	})
}

func GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	user := models.User{}
	result := database.DB.Preload("Country").Where("id = ?", parsedID).First(&user)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Получаем активные заявки пользователя
	var applications []models.GameApplication
	database.DB.
		Preload("Game").
		Preload("User").
		Where("user_id = ? AND is_active = ?", parsedID, true).
		Order("created_at DESC").
		Find(&applications)

	// Скрываем email при просмотре чужого профиля
	currentUserID, _ := utils.GetUserIDFromContext(c)
	if currentUserID != parsedID {
		user.Email = ""
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":         user,
		"applications": applications,
	})
}


func LogoutUser(c *fiber.Ctx) error {
	isProduction := os.Getenv("GO_ENV") == "production"

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   isProduction,
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	// Получаем user ID из cookie
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Проверяем что пользователь редактирует свой профиль
	paramID := c.Params("id")
	if paramID != "" && paramID != userID.String() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only update your own profile",
		})
	}

	type UpdateProfileRequest struct {
		Discord     *string  `json:"discord"`
		Telegram    *string  `json:"telegram"`
		CountryCode *string  `json:"country_code"`
		Description *string  `json:"description"`
		BirthDate   *string  `json:"birth_date"`
		Gender      *string  `json:"gender"`
		Languages   []string `json:"languages"`
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if req.Discord != nil {
		user.Discord = req.Discord
	}
	if req.Telegram != nil {
		user.Telegram = req.Telegram
	}
	if req.CountryCode != nil {
		user.CountryCode = req.CountryCode
	}
	if req.Description != nil {
		user.Description = req.Description
	}
	if req.BirthDate != nil && *req.BirthDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid birth_date format, expected YYYY-MM-DD",
			})
		}
		user.BirthDate = &models.Date{Time: parsed}
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}
	if req.Languages != nil {
		user.Languages = req.Languages
	}

	// Сохраняем изменения
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	// Загружаем обновленные данные с Country
	database.DB.Preload("Country").Where("id = ?", userID).First(&user)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// GetWebSocketToken возвращает токен для WebSocket подключения
// Используется потому что HTTP-only cookie не отправляется на другой порт
func GetWebSocketToken(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Генерируем токен (тот же что и для обычной авторизации)
	token, err := utils.GenerateToken(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
	})
}
