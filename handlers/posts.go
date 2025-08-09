package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kostya-zero/blogger/dto"
	"github.com/kostya-zero/blogger/jwt"
	"github.com/kostya-zero/blogger/models"
	"github.com/kostya-zero/blogger/validation"
	"gorm.io/gorm"
)

type PostsHandler struct {
	DB *gorm.DB
}

func NewPostsHandler(db *gorm.DB) *PostsHandler {
	return &PostsHandler{DB: db}
}

func (ph *PostsHandler) CreatePost(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Access denied.",
		})
	}

	claims, ok := user.(*jwt.TokenClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: bad access token",
		})
	}

	var req dto.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": (*err)[0]})
	}

	newPost := models.Post{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		UserID:      claims.UserID,
		CreatedAt:   time.Now(),
	}

	if err := ph.DB.Create(&newPost).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create new post."})
	}

	return c.JSON(fiber.Map{"success": 1})
}
