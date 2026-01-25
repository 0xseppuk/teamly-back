package main

import (
	"log"

	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/router"
	"github.com/duker221/teamly/internal/services/email"
	"github.com/duker221/teamly/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Инициализация конфигурации JWT
	utils.LoadTokenConfig()

	// Инициализация базы данных
	database.InitDB()

	// Инициализация email сервиса
	if err := email.InitEmailService(); err != nil {
		log.Printf("Warning: Email service initialization failed: %v", err)
	}

	// Создание Fiber приложения
	webApp := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware для логирования
	webApp.Use(logger.New())

	// CORS configuration
	webApp.Use(cors.New(cors.Config{
		AllowOrigins:     utils.GetEnv("CORS_ORIGIN", "http://localhost:3000"),
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Recaptcha-Token",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))

	// Настройка роутов
	router.SetupRoutes(webApp)

	webApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Teamly API",
			"status":  "running",
		})
	})

	// Запуск сервера
	port := utils.GetEnv("PORT", "3001")
	log.Printf("Server starting on port %s", port)
	if err := webApp.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
