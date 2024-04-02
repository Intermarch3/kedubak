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

func CommentRoutes(app *fiber.App, client *mongo.Client, tokenChecker fiber.Handler) {
	comment := app.Group("/comment", tokenChecker, func(c *fiber.Ctx) error {
		return c.Next()
	})
	addComment(client, comment)
}

func addComment(client *mongo.Client, comment fiber.Router) {
	comment.Post("/:id", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		UserId, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "wrong token",
			})
		}

		// Parse request body
		var commentRequest models.Comment
		if err := c.BodyParser(&commentRequest); err != nil || commentRequest.Content == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Bad Request",
			})
		}

		// Get the user
		userCollection := client.Database("keduback").Collection("User")
		objId, _ := primitive.ObjectIDFromHex(UserId)
		user := models.User{}
		_ = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)

		// Create the comment
		comment := models.Comment{
			Content:   commentRequest.Content,
			FirstName: user.FirstName,
			CreatedAt: time.Now(),
			ID:        primitive.NewObjectID().String(),
		}

		// Get the post
		postCollection := client.Database("keduback").Collection("Post")
		postID := c.Params("id")
		objId, _ = primitive.ObjectIDFromHex(postID)
		post := models.Post{}
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Post not found",
			})
		}

		// Add the comment to the post
		post.Comments = append(post.Comments, comment)

		// Update the post
		_, err = postCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": bson.M{"comments": post.Comments}})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"firstName":   user.FirstName,
				"content": comment.Content,
				"createdAt": comment.CreatedAt,
			},
		})
	})
}
