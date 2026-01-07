package handlers

import (
	"time"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterUser(c *fiber.Ctx) error {
	var req models.AuthRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" || req.Nickname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, nickname and password are required",
		})
	}

	// Хэшируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Генерируем токен
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    user,
		"token":   token,
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"user":    user,
		"token":   token,
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":         user,
		"applications": applications,
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	result := database.DB.Preload("Country").Find(&users)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": users,
		"count": len(users),
	})
}

func LogoutUser(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   false,
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

	// Структура для обновления профиля
	type UpdateProfileRequest struct {
		Discord     *string  `json:"discord"`
		Telegram    *string  `json:"telegram"`
		CountryCode *string  `json:"country_code"`
		Description *string  `json:"description"`
		BirthDate   *string  `json:"birth_date"` // Принимаем как строку YYYY-MM-DD
		Gender      *string  `json:"gender"`
		Languages   []string `json:"languages"`
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Находим пользователя
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Обновляем поля напрямую в структуре
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
		// Парсим дату из формата YYYY-MM-DD
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
