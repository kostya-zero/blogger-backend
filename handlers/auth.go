// Package handlers contains handlers for Blogger backend
package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/kostya-zero/blogger/dto"
	"github.com/kostya-zero/blogger/models"
	"github.com/kostya-zero/blogger/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TokenClaims struct {
	UserID uint   `json:"sub"`
	Issuer string `json:"iss"`
	jwt.RegisteredClaims
}

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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": (*err)[0]})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		CreatedAt:    time.Now(),
		PasswordHash: string(hash),
	}

	if err := h.DB.Where("email = ?", user.Email).First(&models.User{}).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user with this email already exists"})
	}

	if err := h.DB.Where("username = ?", user.Username).First(&models.User{}).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user with this username email already exists"})
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

	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": (*err)[0]})
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

	return c.JSON(fiber.Map{"success": 1})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	tok := c.Cookies("access_token", "")

	if tok == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no access token provided"})
	}

	claims, err := h.parseToken(tok, h.Secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	userID := claims.UserID
	var user models.User

	if err := h.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	if user.RefreshToken == nil || *user.RefreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no refresh token found"})
	}

	_, err = h.parseToken(*user.RefreshToken, h.Secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token, " + err.Error()})
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
	tok := c.Cookies("access_token", "")
	claims, err := h.parseToken(tok, h.Secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token is invalid"})
	}

	userID := claims.UserID
	h.DB.Model(&models.User{}).Where("id = ?", userID).Update("refresh_token", "")
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", HTTPOnly: true, Expires: time.Now().Add(-1 * time.Hour)})
	return c.SendStatus(fiber.StatusOK)
}

func (h *AuthHandler) createToken(userID uint, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		UserID: userID,
		Issuer: "blogger",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("could not sign token: " + err.Error())
	}

	return tokenString, nil
}

func (h *AuthHandler) parseToken(tokenStr, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("token malformed")
		}
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("could not parse claims")
	}

	return claims, nil
}
