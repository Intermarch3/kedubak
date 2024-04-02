package main

import (
	"github.com/gofiber/fiber/v2"
	"context"
	"github.com/Taker-Academy/kedubak-Intermarch3/db"
	"github.com/Taker-Academy/kedubak-Intermarch3/jwt"
	"github.com/Taker-Academy/kedubak-Intermarch3/api"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

const jwtSecret = "mdpsecret"


func main() {
	app := fiber.New()
	client := db.ConnectToDb()
	tokenChecker := jwt.NewAuthMiddleware(jwtSecret)

	// Define the routes
	app.Use(cors.New())
	api.UserRoutes(app, client, tokenChecker)
	api.AuthRoutes(app, client, tokenChecker)
	api.PostRoutes(app, client, tokenChecker)

	// Disconnect from the server
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
    app.Listen(":8080")
}
