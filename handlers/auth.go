package handlers

import (
	"blogger/dto"
	"blogger/models"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB            *gorm.DB
	Secret        string
	RefreshSecret string
}

func NewAuthHandler(db *gorm.DB, secret, refreshSecret string) *AuthHandler {
	return &AuthHandler{DB: db, Secret: secret, RefreshSecret: refreshSecret}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		CreatedAt:    time.Now(),
		PasswordHash: string(hash),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create user in database"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "wrong password"})
	}

	access, err := h.createToken(user.ID, h.Secret, 15*time.Minute)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "could not generate access token"})
	}

	refresh, _ := h.createToken(user.ID, h.Secret, 7*24*time.Minute)

	h.DB.Model(&user).Update("refresh_token", refresh)

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    access,
		HTTPOnly: true,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		HTTPOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	return c.JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	tok := c.FormValue("refresh_token")
	claims, err := h.parseToken(tok, h.RefreshSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	userID := uint(claims["sub"].(float64))
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil || user.RefreshToken != &tok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	newAccess, _ := h.createToken(userID, h.Secret, 15*time.Minute)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccess,
		HTTPOnly: true,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	return c.JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	tok := c.FormValue("refresh_token")
	claims, err := h.parseToken(tok, h.RefreshSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	userID := uint(claims["sub"].(float64))
	h.DB.Model(&models.User{}).Where("id = ?", userID).Update("refresh_token", "")
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", HTTPOnly: true, Expires: time.Now().Add(-1 * time.Hour)})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", HTTPOnly: true, Expires: time.Now().Add(-1 * time.Hour)})
	return c.SendStatus(fiber.StatusOK)
}

func (h *AuthHandler) createToken(userID uint, secret string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iss": "blogger",
		"exp": time.Now().Add(ttl).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (h *AuthHandler) parseToken(tokenStr, secret string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return tok.Claims.(jwt.MapClaims), nil
}
