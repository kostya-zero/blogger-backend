// Package handlers contains handlers for Blogger backend
package handlers

import (
	"time"

	"github.com/kostya-zero/blogger/dto"
	"github.com/kostya-zero/blogger/jwt"
	"github.com/kostya-zero/blogger/models"
	"github.com/kostya-zero/blogger/validation"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB     *gorm.DB
	Secret string
}

func NewAuthHandler(db *gorm.DB, secret string) *AuthHandler {
	return &AuthHandler{DB: db, Secret: secret}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": (*err)[0]})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not hash password"})
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		CreatedAt:    time.Now(),
		PasswordHash: string(hash),
	}

	if err := h.DB.Where("email = ?", user.Email).First(&models.User{}).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User with this email already exists"})
	}

	if err := h.DB.Where("username = ?", user.Username).First(&models.User{}).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User with this username email already exists"})
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user in database"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": (*err)[0]})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Wrong password"})
	}

	access, err := jwt.CreateToken(user.ID, h.Secret, 15*time.Minute)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Could not generate access token"})
	}

	refresh, _ := jwt.CreateToken(user.ID, h.Secret, 7*24*time.Minute)

	h.DB.Model(&user).Update("refresh_token", refresh)

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    access,
		HTTPOnly: true,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	return c.JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	tok := c.Cookies("access_token", "")

	if tok == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No access token provided"})
	}

	claims, err := jwt.ParseToken(tok, h.Secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	userID := claims.UserID
	var user models.User

	if err := h.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	if user.RefreshToken == nil || *user.RefreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No refresh token found"})
	}

	_, err = jwt.ParseToken(*user.RefreshToken, h.Secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token, " + err.Error()})
	}

	newAccess, _ := jwt.CreateToken(userID, h.Secret, 15*time.Minute)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccess,
		HTTPOnly: true,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	return c.JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	tok := c.Cookies("access_token", "")
	claims, err := jwt.ParseToken(tok, h.Secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token is invalid"})
	}

	userID := claims.UserID
	h.DB.Model(&models.User{}).Where("id = ?", userID).Update("refresh_token", "")
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", HTTPOnly: true, Expires: time.Now().Add(-1 * time.Hour)})
	return c.SendStatus(fiber.StatusOK)
}
