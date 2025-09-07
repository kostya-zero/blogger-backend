package routes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kostya-zero/blogger/helpers"
	"github.com/kostya-zero/blogger/models"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{DB: db}
}

func (sh *SettingsHandler) UpdateDisplayName(c *fiber.Ctx) error {
	claims, err := helpers.GetClaimsFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userName := c.Query("username")

	if userName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'username' parameter is required"})
	}

	var user models.User
	result := sh.DB.Where("id = ?", claims.UserID).First(&user)
	// Not possible, but anyways
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found (idk how)."})
	}

	// Search if user with the same username exists
	var count int64
	sh.DB.Model(&models.User{}).Where("username = ?", userName).Count(&count)
	fmt.Printf("%d\n", count)
	if count > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User with the same username exists."})
	}

	user.Username = userName
	if err := sh.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update username"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": 1})
}
