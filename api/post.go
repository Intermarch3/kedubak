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
	getPostById(client, post)
	deletePostById(client, post)
	addVote(client, post)
}

func addVote(client *mongo.Client, post fiber.Router) {
	post.Post("/vote/:id", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		UserId, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		// Check if the user has already voted in the last minute
		lastTime := getUserVoteTime(client, UserId)
		if lastTime.Add(time.Minute * 1).After(time.Now()) {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"ok": false,
				"error": "You can only vote once per minute",
			})
		}

		postCollection := client.Database("keduback").Collection("Post")
		postID := c.Params("id")
		objId, err := primitive.ObjectIDFromHex(postID)
		if err != nil {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
				"ok": false,
				"error": "Invalid ID",
			})
		}
		//get the post
		post := models.Post{}
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok": false,
				"error": "Post not found",
			})
		}

		// Check if the user has already voted
		for _, vote := range post.UpVotes {
			if vote == UserId {
				return c.Status(http.StatusConflict).JSON(fiber.Map{
					"ok": false,
					"error": "Already voted for this post",
				})
			}
		}

		// Update the user vote time
		updateUserVoteTime(client, UserId)

		// Add the user to the upvotes
		post.UpVotes = append(post.UpVotes, UserId)

		// Update the post
		_, err = postCollection.UpdateOne(context.Background(), bson.M{"_id": objId}, bson.M{"$set": bson.M{"upVotes": post.UpVotes}})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"message": "post upvoted",
		})
	})
}

func deletePostById(client *mongo.Client, post fiber.Router) {
	post.Delete("/:id", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		UserId, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		postCollection := client.Database("keduback").Collection("Post")
		postID := c.Params("id")
		objId, err := primitive.ObjectIDFromHex(postID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok": false,
				"error": "Invalid ID",
			})
		}
		//get the post
		post := models.Post{}
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"ok": false,
				"error": "Post not found",
			})
		}

		// Check if the user is the owner of the post
		if post.UserId != UserId {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"ok": false,
				"error": "user not the owner of the post",
			})
		}

		//delete post
		_, err = postCollection.DeleteOne(context.Background(), bson.M{"_id": objId})
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"ok": false,
					"error": "Post not found",
				})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		// change nil array to empty array
		if (post.UpVotes == nil) {
			post.UpVotes = []string{}
		}
		if (post.Comments == nil) {
			post.Comments = []models.Comment{}
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": fiber.Map{
				"createdAt": post.CreatedAt,
				"userId": post.UserId,
				"firstName": post.FirstName,
				"title": post.Title,
				"content": post.Content,
				"comments": post.Comments,
				"upVotes": post.UpVotes,
				"removed": true,
			},
		})
	})
}

func getPostById(client *mongo.Client, post fiber.Router) {
	post.Get("/:id", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		_, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		postCollection := client.Database("keduback").Collection("Post")
		postID := c.Params("id")
		objId, err := primitive.ObjectIDFromHex(postID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"ok": false,
				"error": "Invalid ID",
			})
		}

		//get the post
		post := models.Post{}
		err = postCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&post)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{
					"ok": false,
					"error": "Post not found",
				})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}

		// change nil array to empty array
		if (post.UpVotes == nil) {
			post.UpVotes = []string{}
		}
		if (post.Comments == nil) {
			post.Comments = []models.Comment{}
		}

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"ok": true,
			"data": post,
		})
	})
}

func GetMyPosts(client *mongo.Client, post fiber.Router) {
	post.Get("/me", func(c *fiber.Ctx) error {
		// Get the token from the header Authorization
		token := c.Get("Authorization")
		userID, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
				"error": "wrong token",
			})
		}

		postCollection := client.Database("keduback").Collection("Post")
		//get all posts of the user
		cursor, err := postCollection.Find(context.Background(), bson.M{"userId": userID})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
				"error": "Internal Server Error",
			})
		}
		posts := []models.Post{}
		if err = cursor.All(context.Background(), &posts); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
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
				"ok": false,
				"error": "Bad Request",
			})
		}

		// Get the token from the header Authorization
		token := c.Get("Authorization")
		userID, err := jwt.GetUserID(token, client)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"ok": false,
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
		// ajout d'une reponse en cas d'erreur (non noter sur la doc)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
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
				"ok": false,
				"error": "Internal Server Error",
			})
		}
		posts := []models.Post{}
		if err = cursor.All(context.Background(), &posts); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ok": false,
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
