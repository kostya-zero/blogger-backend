package main

import (
	"blogger/handlers"
	"blogger/models"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	secret := "supersecret"
	refreshSecret := "refreshsecret"

	dsn := "host=localhost user=blogger password=blogger dbname=blogger port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to open connection to database: %s", err.Error())
		os.Exit(1)
	}

	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Like{})

	ah := handlers.NewAuthHandler(db, secret, refreshSecret)
	uh := handlers.NewUserHandler(db)

	app := fiber.New()

	// Auth group
	authGroup := app.Group("/auth")
	authGroup.Post("/register", ah.Register)
	authGroup.Post("/login", ah.Login)
	authGroup.Post("/refresh", ah.Refresh)
	authGroup.Post("/logout", ah.Logout)

	// User group
	userGroup := app.Group("/user")
	userGroup.Get("/get", uh.GetUser)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Listen(":3000")
}
