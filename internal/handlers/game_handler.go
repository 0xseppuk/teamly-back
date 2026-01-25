package handlers

import (
	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetAllGames(c *fiber.Ctx) error {
	var games []models.Game
	search := c.Query("search")

	query := database.DB

	// Add search filter if provided
	if search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	result := query.Find(&games)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch games",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"games": games,
		"count": len(games),
	})
}

func GetGameById(c *fiber.Ctx) error {
	gameID := c.Params("id")
	if gameID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Game ID is required",
		})
	}

	parsedID, err := uuid.Parse(gameID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid game ID format",
		})
	}

	game := models.Game{}
	result := database.DB.Where("id = ?", parsedID).First(&game)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"game": game,
	})

}
