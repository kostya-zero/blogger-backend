package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kostya-zero/blogger/jwt"
	"github.com/kostya-zero/blogger/models"
	"github.com/kostya-zero/blogger/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	println("Starting Blogger Backend...")
	println("Loading dotenv...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	secret := os.Getenv("BLOGGER_JWT_SECRET")
	dsn := os.Getenv("BLOGGER_GORM_DATABASE_STRING")

	println("Connecting to database...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to open connection to database: %s", err.Error())
		os.Exit(1)
	}
	println("Successfully connected to database.")

	println("Running migrations...")
	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Like{})

	println("Setting up Fiber...")
	ah := routes.NewAuthHandler(db, secret)
	uh := routes.NewUserHandler(db)
	ph := routes.NewPostsHandler(db)
	sh := routes.NewSettingsHandler(db)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	app.Use(logger.New())

	// Auth group
	authGroup := app.Group("/auth")
	authGroup.Post("/register", ah.Register)
	authGroup.Post("/login", ah.Login)
	authGroup.Post("/refresh", ah.Refresh)
	authGroup.Post("/logout", ah.Logout)

	// Users group
	usersGroup := app.Group("/users")
	usersGroup.Get("/get", uh.GetUser)
	usersGroup.Get("/getLikes", uh.GetLikes)
	usersGroup.Get("/getPosts", uh.GetUsersPosts)

	postsGroup := app.Group("/posts")
	postsGroup.Post("/create", jwt.JwtMiddleware(secret), ph.CreatePost)
	postsGroup.Get("/get", ph.GetPost)
	postsGroup.Post("/like", jwt.JwtMiddleware(secret), ph.Like)

	settingsGroup := app.Group("/settings")
	settingsGroup.Post("/update-username", jwt.JwtMiddleware(secret), sh.UpdateUserName)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	println("Running Fiber App...")
	app.Listen(":3000")
}
