package router

import (
	"github.com/duker221/teamly/internal/handlers"
	"github.com/duker221/teamly/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Общий rate limiter для всех API запросов (100 req/min)
	api.Use(middleware.APIRateLimiter())

	//auth
	auth := api.Group("/auth")
	// Строгий лимит для логина и регистрации (5 req/min) - защита от брутфорса
	auth.Post("/login", middleware.AuthRateLimiter(), middleware.RecaptchaMiddleware(), handlers.LoginUser)
	auth.Post("/register", middleware.AuthRateLimiter(), middleware.RecaptchaMiddleware(), handlers.RegisterUser)
	auth.Post("/logout", handlers.LogoutUser)
	auth.Get("/me", handlers.GetMe)
	auth.Get("/ws-token", handlers.GetWebSocketToken) // Токен для WebSocket
	auth.Patch("/me", handlers.UpdateProfile)
	// Password reset endpoints
	auth.Post("/forgot-password", middleware.AuthRateLimiter(), middleware.RecaptchaMiddleware(), handlers.ForgotPassword)
	auth.Post("/reset-password", middleware.AuthRateLimiter(), handlers.ResetPassword)

	//users
	users := api.Group("/users")
	users.Get("/", handlers.GetAllUsers)
	users.Get("/:id", handlers.GetUserByID)
	users.Get("/:id/applications", handlers.GetApplicationsByUserID)
	users.Patch("/:id", handlers.UpdateProfile)

	//countries
	countries := api.Group("/countries")
	countries.Get("/", handlers.GetAllCountries)
	countries.Get("/region/:region", handlers.GetCountriesByRegion)

	//games
	games := api.Group("/games")
	games.Get("/", handlers.GetAllGames)
	games.Get("/:id", handlers.GetGameById)

	//game applications
	applications := api.Group("/applications")
	applications.Get("/", handlers.GetAllApplications)
	applications.Get("/my", middleware.AuthRequired, handlers.GetUserApplications)
	applications.Get("/:id", handlers.GetApplicationByID)
	applications.Post("/", middleware.AuthRequired, middleware.CreateApplicationRateLimiter(), handlers.CreateGameApplication)
	applications.Patch("/:id", middleware.AuthRequired, handlers.UpdateApplication)
	applications.Delete("/:id", middleware.AuthRequired, handlers.DeleteApplication)

	// Application responses
	applications.Post("/:id/responses", middleware.AuthRequired, handlers.CreateApplicationResponse)
	applications.Get("/:id/responses", middleware.AuthRequired, handlers.GetApplicationResponses)

	//responses
	responses := api.Group("/responses")
	responses.Get("/my", middleware.AuthRequired, handlers.GetMyResponses)
	responses.Patch("/:id", middleware.AuthRequired, handlers.UpdateResponseStatus)

	// Conversations & Messages
	conversations := api.Group("/conversations", middleware.AuthRequired)
	conversations.Get("/", handlers.GetUserConversations)                // List all user's conversations
	conversations.Get("/unread-count", handlers.GetUnreadCount)          // Get total unread count
	conversations.Get("/:id", handlers.GetConversationByID)              // Get specific conversation
	conversations.Get("/:id/messages", handlers.GetConversationMessages) // Get messages with pagination
	conversations.Patch("/:id/read", handlers.MarkMessagesAsRead)        // Mark all messages as read
}
