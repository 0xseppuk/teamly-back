package handlers

import (
	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetAllGames(c *fiber.Ctx) error {
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 8)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 8
	}

	offset := (page - 1) * limit

	var total int64
	countQuery := database.DB.Model(&models.Game{})

	if search != "" {
		countQuery = countQuery.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count games",
		})
	}

	var games []models.Game
	dataQuery := database.DB.Model(&models.Game{})

	if search != "" {
		dataQuery = dataQuery.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	if err := dataQuery.
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&games).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch games",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"games": games,
		"count": len(games),
		"total": total,
	})
}

func GetGameBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Game slug is required",
		})
	}

	game := models.Game{}
	result := database.DB.Where("slug = ?", slug).First(&game)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(game)
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
