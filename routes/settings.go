package routes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kostya-zero/blogger/helpers"
	"github.com/kostya-zero/blogger/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{DB: db}
}

func (sh *SettingsHandler) UpdateUserName(c *fiber.Ctx) error {
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

func (sh *SettingsHandler) UpdateDisplayName(c *fiber.Ctx) error {
	claims, err := helpers.GetClaimsFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	displayName := c.Query("displayName")
	if displayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'displayName' parameter is required"})
	}

	var user models.User
	result := sh.DB.Where("id = ?", claims.UserID).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found (idk how)."})
	}

	user.DisplayName = &displayName
	if err := sh.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update display name"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": 1})
}

func (sh *SettingsHandler) UpdatePassword(c *fiber.Ctx) error {
	claims, err := helpers.GetClaimsFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	newPassword := c.Query("password")
	if newPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'password' parameter is required"})
	}

	var user models.User
	result := sh.DB.Where("id = ?", claims.UserID).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	user.PasswordHash = string(hash)
	if err := sh.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": 1})
}
