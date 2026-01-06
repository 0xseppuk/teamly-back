package handlers

import (
	"github.com/duker221/teamly/internal/database"
	"github.com/duker221/teamly/internal/models"
	"github.com/gofiber/fiber/v2"
)

// GetAllCountries returns all available countries
func GetAllCountries(c *fiber.Ctx) error {
	var countries []models.Country
	result := database.DB.Find(&countries)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch countries",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"countries": countries,
		"count":     len(countries),
	})
}

// GetCountriesByRegion returns countries filtered by region
func GetCountriesByRegion(c *fiber.Ctx) error {
	region := c.Params("region")
	if region == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Region is required",
		})
	}

	var countries []models.Country
	result := database.DB.Where("region = ?", region).Find(&countries)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch countries",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"countries": countries,
		"count":     len(countries),
		"region":    region,
	})
}
