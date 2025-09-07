package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{DB: db}
}

func (sh *SettingsHandler) UpdateDisplayName(c *fiber.Ctx) error {
	userName := c.Query("username")

	if userName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'username' parameter is required"})
	}

	return c.Status(fiber.StatusOK).SendString("OK")
}
