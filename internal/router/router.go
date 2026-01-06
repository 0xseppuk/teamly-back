package router

import (
	"github.com/duker221/teamly/internal/handlers"
	"github.com/duker221/teamly/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	//auth
	auth := api.Group("/auth")
	auth.Post("/login", handlers.LoginUser)
	auth.Post("/register", handlers.RegisterUser)
	auth.Post("/logout", handlers.LogoutUser)
	auth.Get("/me", handlers.GetMe)
	auth.Patch("/me", handlers.UpdateProfile)

	//users
	users := api.Group("/users")
	users.Get("/", handlers.GetAllUsers)
	users.Get("/:id", handlers.GetUserByID)
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
	applications.Post("/", middleware.AuthRequired, handlers.CreateGameApplication)
	applications.Patch("/:id", middleware.AuthRequired, handlers.UpdateApplication)
	applications.Delete("/:id", middleware.AuthRequired, handlers.DeleteApplication)
}
