package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Taker-Academy/kedubak-Intermarch3/jwt"
	"github.com/Taker-Academy/kedubak-Intermarch3/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthRoutes(app *fiber.App, client *mongo.Client, tokenChecker fiber.Handler) {
	auth := app.Group("/auth", func(c *fiber.Ctx) error {
		return c.Next()
	})
	Register(client, auth)
	Login(client, auth)
}

func Login(client *mongo.Client, auth fiber.Router) {
	auth.Post("/login", func(c *fiber.Ctx) error {
		// Parse request body
		var loginRequest models.User
		if err := c.BodyParser(&loginRequest); err != nil || loginRequest.Email == "" || loginRequest.Password == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok": false,
				"error": "Bad Request",
			})
		}

		userCollection := client.Database("keduback").Collection("User")

		// get user from db
		existingUser := models.User{}
		err := userCollection.FindOne(context.Background(), bson.M{"email": loginRequest.Email}).Decode(&existingUser)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong credentials",
			})
		}

		// Check if the password is correct
		if !CheckPasswordHash(loginRequest.Password, existingUser.Password) {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong credentials",
			})
		}

		// Generate JWT token
		userID := existingUser.ID.Hex()
		token := jwt.GetToken(userID)
		if token == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"token": token,
				"user": fiber.Map{
					"email":     existingUser.Email,
					"firstName": existingUser.FirstName,
					"lastName":  existingUser.LastName,
				},
			},
		})
	})
}

func Register(client *mongo.Client, auth fiber.Router) {
	auth.Post("/register", func(c *fiber.Ctx) error {
		// Parse request body
		var newUser models.User
		if err := c.BodyParser(&newUser); err != nil || newUser.Email == "" ||
			newUser.Password == "" || newUser.FirstName == "" || newUser.LastName == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok": false,
				"error": "Bad Request",
			})
		}

		userCollection := client.Database("keduback").Collection("User")

		// Check if user with the same email already exists
		existingUser := models.User{}
		err := userCollection.FindOne(context.Background(), bson.M{"email": newUser.Email}).Decode(&existingUser)
		if err == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "User with the same email already exists",
			})
		}

		// Hash the password
		hash, err := HashPassword(newUser.Password)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}
		newUser.Password = hash

		// Insert user into the database
		newUser.CreatedAt = time.Now()
		newUser.LastUpVote = time.Now().Add(-1 * time.Minute)

		// save new user in db
		res, err := userCollection.InsertOne(context.Background(), newUser)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		// Generate JWT token
		userID := res.InsertedID.(primitive.ObjectID).Hex()
		token := jwt.GetToken(userID)
		if token == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"token": token,
				"user": fiber.Map{
					"email":     newUser.Email,
					"firstName": newUser.FirstName,
					"lastName":  newUser.LastName,
				},
			},
		})
	})
}