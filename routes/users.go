package routes

import (
	"errors"
	"strconv"

	"github.com/kostya-zero/blogger/models"

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

func (uh *UserHandler) GetLikes(c *fiber.Ctx) error {
	userIDStr := c.Query("id", "")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The 'id' parameter is required"})
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var likes []models.Like
	result := uh.DB.Preload("Post").Find(&likes, "user_id = ?", userID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	if len(likes) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Nothing found."})
	}

	var posts []models.Post
	for _, v := range likes {
		posts = append(posts, v.Post)
	}

	return c.Status(fiber.StatusOK).JSON(posts)
}
