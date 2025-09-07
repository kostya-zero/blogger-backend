package routes

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kostya-zero/blogger/dto"
	"github.com/kostya-zero/blogger/helpers"
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
	claims, err := helpers.GetClaimsFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
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

func (ph *PostsHandler) GetPost(c *fiber.Ctx) error {
	postID := c.Query("id", "")

	if postID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The 'id' parameter is required"})
	}

	var post models.Post
	if err := ph.DB.Preload("User").Where("id = ?", postID).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not retrieve user data"})
	}

	return c.JSON(post)
}

func (ph *PostsHandler) Like(c *fiber.Ctx) error {
	claims, err := helpers.GetClaimsFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	postID := c.Query("id", "")

	if postID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The 'id' parameter is required"})
	}

	postIntID, err := strconv.Atoi(postID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad post ID"})
	}

	if err = ph.DB.Where("id = ?", postIntID).First(&models.Post{}).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	like := models.Like{
		UserID: claims.UserID,
		PostID: uint(postIntID),
	}

	if err = ph.DB.Create(like).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to like the post"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"succes": 1})
}
