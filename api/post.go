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

func PostRoutes(app *fiber.App, client *mongo.Client, tokenChecker fiber.Handler) {
	post := app.Group("/post", tokenChecker, func(c *fiber.Ctx) error {
		return c.Next()
	})
	GetPosts(client, post)
	CreatePost(client, post)
	GetMyPosts(client, post)
}

func GetMyPosts(client *mongo.Client, post fiber.Router) {
	post.Get("/me", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		userID, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "wrong token",
			})
		}

		postCollection := client.Database("keduback").Collection("Post")
		//get all posts of the user
		cursor, err := postCollection.Find(context.Background(), bson.M{"userId": userID})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
		posts := []models.Post{}
		if err = cursor.All(context.Background(), &posts); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		// change nil array to empty array
		for i, post := range posts {
			if (post.UpVotes == nil) {
				posts[i].UpVotes = []string{}
			}
			if (post.Comments == nil) {
				posts[i].Comments = []models.Comment{}
			}
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": posts,
		})
	})
}

func CreatePost(client *mongo.Client, post fiber.Router) {
	post.Post("/", func(c *fiber.Ctx) error {
		// Parse request body
		var postRequest models.Post
		if err := c.BodyParser(&postRequest); err != nil || postRequest.Title == "" || postRequest.Content == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Bad Request",
			})
		}

		// Get the token from the header Authorization
		token := c.Get("Authorization")
		userID, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "wrong token",
			})
		}

		userCollection := client.Database("keduback").Collection("User")
		objId, _ := primitive.ObjectIDFromHex(userID)
		//get the user
		user := models.User{}
		_ = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)

		postCollection := client.Database("keduback").Collection("Post")

		// Create a new post
		newPost := models.Post{
			CreatedAt: time.Now(),
			UserId:    userID,
			FirstName: user.FirstName,
			Title:     postRequest.Title,
			Content:   postRequest.Content,
			Comments:  []models.Comment{},
			UpVotes:   []string{},
		}

		_, err = postCollection.InsertOne(context.Background(), newPost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"ok": true,
			"data": newPost,
		})
	})
}

func GetPosts(client *mongo.Client, post fiber.Router) {
	post.Get("/", func(c *fiber.Ctx) error {
		postCollection := client.Database("keduback").Collection("Post")
		//get all posts
		cursor, err := postCollection.Find(context.Background(), bson.M{})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}
		posts := []models.Post{}
		if err = cursor.All(context.Background(), &posts); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		// change nil array to empty array
		for i, post := range posts {
			if (post.UpVotes == nil) {
				posts[i].UpVotes = []string{}
			}
			if (post.Comments == nil) {
				posts[i].Comments = []models.Comment{}
			}
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": posts,
		})
	})
}
