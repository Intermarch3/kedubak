package main

import (
	"github.com/gofiber/fiber/v2"
	"context"
	"fmt"
	"os"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connect_to_db() *mongo.Client {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	// Disconnect from the server
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client
}

func main() {
	app := fiber.New()
	client := connect_to_db()
	// Define your routes
	app.Get("/", func(c *fiber.Ctx) error {
		fmt.Println("get / called")
		return c.SendString("Hello, Coders! Welcome to Go programming language.")
	})
    app.Listen(":8080")
}