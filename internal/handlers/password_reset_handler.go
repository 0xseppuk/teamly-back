package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/duker221/teamly/internal/services/email"
	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	tokenLength    = 32 // 32 bytes = 64 hex characters
	tokenExpiresIn = 1 * time.Hour
)

// ForgotPassword обрабатывает запрос на восстановление пароля
// POST /api/auth/forgot-password
func ForgotPassword(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ForgotPassword] Body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Всегда возвращаем успех для защиты от email enumeration
	successResponse := fiber.Map{
		"message": "If an account with this email exists, a password reset link has been sent",
	}

	// Ищем пользователя
	var user models.User
	result := database.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		log.Printf("[ForgotPassword] User not found for email: %s", req.Email)
		// Возвращаем такой же ответ, чтобы не раскрывать существование email
		return c.Status(fiber.StatusOK).JSON(successResponse)
	}

	// Инвалидируем все предыдущие неиспользованные токены для этого пользователя
	database.DB.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", user.ID).
		Update("used_at", time.Now())

	// Генерируем новый токен
	rawToken, err := generateSecureToken(tokenLength)
	if err != nil {
		log.Printf("[ForgotPassword] Failed to generate token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate reset token",
		})
	}

	// Хешируем токен для хранения в БД
	tokenHash := hashToken(rawToken)

	// Создаем запись токена
	resetToken := models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(tokenExpiresIn),
	}

	if err := database.DB.Create(&resetToken).Error; err != nil {
		log.Printf("[ForgotPassword] Failed to save token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create reset token",
		})
	}

	// Отправляем email с raw токеном (не хешем)
	if err := email.SendPasswordResetEmail(user.Email, rawToken); err != nil {
		log.Printf("[ForgotPassword] Failed to send email: %v", err)
		// Не возвращаем ошибку клиенту для защиты от enumeration
	}

	log.Printf("[ForgotPassword] Reset token created for user: %s", user.ID)
	return c.Status(fiber.StatusOK).JSON(successResponse)
}

// ResetPassword обрабатывает сброс пароля
// POST /api/auth/reset-password
func ResetPassword(c *fiber.Ctx) error {
	var req models.ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ResetPassword] Body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}

	if req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New password is required",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters",
		})
	}

	// Хешируем переданный токен для поиска в БД
	tokenHash := hashToken(req.Token)

	// Ищем токен в БД
	var resetToken models.PasswordResetToken
	result := database.DB.Where("token = ?", tokenHash).First(&resetToken)
	if result.Error != nil {
		log.Printf("[ResetPassword] Token not found")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Проверяем валидность токена
	if !resetToken.IsValid() {
		log.Printf("[ResetPassword] Token is invalid (expired or used)")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Хешируем новый пароль
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("[ResetPassword] Failed to hash password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	// Обновляем пароль пользователя
	if err := database.DB.Model(&models.User{}).
		Where("id = ?", resetToken.UserID).
		Update("password_hash", hashedPassword).Error; err != nil {
		log.Printf("[ResetPassword] Failed to update password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	// Помечаем токен как использованный
	now := time.Now()
	resetToken.UsedAt = &now
	database.DB.Save(&resetToken)

	log.Printf("[ResetPassword] Password reset successfully for user: %s", resetToken.UserID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}

// generateSecureToken генерирует криптографически безопасный токен
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken хеширует токен с помощью SHA256
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
