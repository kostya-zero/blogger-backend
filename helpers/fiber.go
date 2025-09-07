package helpers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/kostya-zero/blogger/jwt"
)

func GetClaimsFromContext(c *fiber.Ctx) (*jwt.TokenClaims, error) {
	user := c.Locals("user")
	if user == nil {
		return nil, errors.New("Access denied.")
	}

	claims, ok := user.(*jwt.TokenClaims)
	if !ok {
		return nil, errors.New("Unauthorized: bad access token")
	}

	return claims, nil
}
