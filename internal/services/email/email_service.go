package email

import (
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

var client *resend.Client
var fromEmail string
var frontendURL string

// InitEmailService инициализирует email сервис с Resend API
func InitEmailService() error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("Warning: RESEND_API_KEY not set, email service disabled")
		return nil
	}

	client = resend.NewClient(apiKey)
	fromEmail = getEnvOrDefault("EMAIL_FROM", "Teamly <noreply@teamly.app>")
	frontendURL = getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")

	log.Println("Email service initialized with Resend")
	return nil
}

// IsEnabled проверяет, включен ли email сервис
func IsEnabled() bool {
	return client != nil
}

// SendPasswordResetEmail отправляет email с ссылкой для сброса пароля
func SendPasswordResetEmail(toEmail, token string) error {
	if !IsEnabled() {
		log.Printf("Email service disabled, would send password reset to: %s", toEmail)
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)
	htmlContent := GetPasswordResetEmailHTML(resetURL)
	textContent := GetPasswordResetEmailText(resetURL)

	params := &resend.SendEmailRequest{
		From:    fromEmail,
		To:      []string{toEmail},
		Subject: "Сброс пароля - Teamly",
		Html:    htmlContent,
		Text:    textContent,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Password reset email sent to %s, ID: %s", toEmail, sent.Id)
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
