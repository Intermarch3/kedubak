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

func UserRoutes(app *fiber.App, client *mongo.Client, tokenChecker fiber.Handler) {
	user := app.Group("/user", tokenChecker, func(c *fiber.Ctx) error {
		return c.Next()
	})
	Me(client, user)
	Edit(client, user)
	Remove(client, user)
}

func Remove(client *mongo.Client, user fiber.Router) {
	user.Delete("/remove", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		id, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		userCollection := client.Database("keduback").Collection("User")
		objId, _ := primitive.ObjectIDFromHex(id)
		//get the user
		user := models.User{}
		err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok": false,
				"error": "User not found",
			})
		}

		//delete user
		_, err = userCollection.DeleteOne(context.Background(), bson.M{"_id": objId})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"removed":   true,
			},
		})
	})
}

func Edit(client *mongo.Client, user fiber.Router) {
	user.Put("/edit", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		id, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		// Parse request body
		var newUser models.User
		if err := c.BodyParser(&newUser); err != nil || (newUser.Email == "" &&
			newUser.FirstName == "" && newUser.LastName == "" && newUser.Password == "") {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
				"ok": false,
				"error": "unable to validate the request body",
			})
		}

		userCollection := client.Database("keduback").Collection("User")

		// Find the user by ID
		user := models.User{}
		objId, _ := primitive.ObjectIDFromHex(id)
		_ = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)

		// Update the user
		if (newUser.Email != "") {
			user.Email = newUser.Email
		}
		if (newUser.FirstName != "") {
			user.FirstName = newUser.FirstName
		}
		if (newUser.LastName != "") {
			user.LastName = newUser.LastName
		}
		if (newUser.Password != "") {
			hash, err := HashPassword(newUser.Password)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"ok": false,
					"error": "Internal Server Error",
				})
			}
			user.Password = hash
		}

		_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": user})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
			},
		})
	})
}

func Me(client *mongo.Client, user fiber.Router) {
	user.Get("/me", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		id, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		userCollection := client.Database("keduback").Collection("User")
		// Find the user by ID
		user := models.User{}
		
		objId, _ := primitive.ObjectIDFromHex(id)
		err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "internal server error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
			},
		})
	})
}

func getUserVoteTime(client *mongo.Client, id string) time.Time {
	userCollection := client.Database("keduback").Collection("User")
	objId, _ := primitive.ObjectIDFromHex(id)
	user := models.User{}
	err := userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
	if err != nil {
		return time.Now()
	} else {
		return user.LastUpVote
	}
}

func updateUserVoteTime(client *mongo.Client, id string) {
	userCollection := client.Database("keduback").Collection("User")
	objId, _ := primitive.ObjectIDFromHex(id)
	user := models.User{}
	user.LastUpVote = time.Now()
	userCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": user})
}