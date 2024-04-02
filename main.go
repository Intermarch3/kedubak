package main

import (
	"fmt"
	"keduback/db"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	client := db.ConnectToDb()
	// Define your routes
	app.Get("/", func(c *fiber.Ctx) error {
		fmt.Println("get / called")
		return c.SendString("Hello, Coders! Welcome to Go programming language.")
	})
	fmt.Println(client.NumberSessionsInProgress())
    app.Listen(":8080")
}