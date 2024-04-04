package main

import (
	"github.com/gofiber/fiber/v2"
	"context"
	"os"
	"github.com/Taker-Academy/kedubak-Intermarch3/db"
	"github.com/Taker-Academy/kedubak-Intermarch3/jwt"
	"github.com/Taker-Academy/kedubak-Intermarch3/api"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	secret := os.Getenv("SECRET_STR")
	if secret == "" {
		panic("SECRET_STR is not set")
	}
	app := fiber.New()
	client := db.ConnectToDb()
	tokenChecker := jwt.NewAuthMiddleware(secret)

	// Define the routes
	app.Use(cors.New())
	api.UserRoutes(app, client, tokenChecker)
	api.AuthRoutes(app, client, tokenChecker)
	api.PostRoutes(app, client, tokenChecker)
	api.CommentRoutes(app, client, tokenChecker)

	// Disconnect from the server
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
    app.Listen(":8080")
}
