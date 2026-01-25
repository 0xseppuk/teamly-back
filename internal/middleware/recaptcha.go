package middleware

import (
	"log"

	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func RecaptchaMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("X-Recaptcha-Token")
		log.Println(token)
		if token == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "recaptcha token missing"})
		}

		ok, err := utils.VerifyRecaptcha(token)
		log.Println(ok, err)
		if err != nil || !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "recaptcha failed"})
		}

		return c.Next()
	}
}
