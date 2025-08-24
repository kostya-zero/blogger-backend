package handlers

import (
	"errors"

	"github.com/kostya-zero/blogger/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (uh *UserHandler) GetUser(c *fiber.Ctx) error {
	userID := c.Query("id")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'id' parameter is required"})
	}

	var user models.User

	if err := uh.DB.Where("username = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not retrieve user data"})
	}

	return c.JSON(user)
}

func (uh *UserHandler) GetUsersLikes(c *fiber.Ctx) error {
	userID := c.Query("id", "")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'id' parameter is required"})
	}

	var user models.User

	if err := uh.DB.Where("username = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not retrieve user data"})
	}

	if len(user.Likes) == 0 {
		return c.JSON(fiber.Map{})
	}

	return c.JSON(user.Likes)
}

func (uh *UserHandler) GetUsersPosts(c *fiber.Ctx) error {
	userID := c.Query("id", "")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "'id' parameter is required"})
	}

	var user models.User

	if err := uh.DB.Where("username = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not retrieve user data"})
	}

	if len(user.Posts) == 0 {
		return c.JSON(fiber.Map{})
	}

	return c.JSON(user.Posts)
}
