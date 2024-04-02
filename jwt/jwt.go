package jwt

import (
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/gofiber/fiber/v2"
	"github.com/Taker-Academy/kedubak-Intermarch3/models"
	jwtware "github.com/gofiber/jwt/v3"
	jtoken "github.com/golang-jwt/jwt/v4"
)

const jwtSecret = "mdpsecret"

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GetSecret() string {
	return jwtSecret
}

func GetClaims(tokenString string) (*jtoken.Token, error) {
	token, err := jtoken.Parse(tokenString, func(token *jtoken.Token) (interface{}, error) {
		return []byte(GetSecret()), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func GetUserID(tokenString string, client *mongo.Client) (string, error) {
	tokenString = tokenString[7:] // Remove the Bearer prefix
	token, err := GetClaims(tokenString)
	if err != nil {
		return "", err
	}
	claims := token.Claims.(jtoken.MapClaims)

	//check if userId exist in the db
	userCollection := client.Database("keduback").Collection("User")
	objId, _ := primitive.ObjectIDFromHex(claims["ID"].(string))
	user := models.User{}
	err = userCollection.FindOne(context.Background(), bson.M{"_id": objId}).Decode(&user)
	if err != nil {
		return "", err
	}
	return claims["ID"].(string), nil
}

func GetToken(userID string) string {
	// Create the JWT claims, which includes the user ID and expiry time
	claims := jtoken.MapClaims{
		"ID":    userID,
		"exp":   time.Now().Add(time.Hour * 24 * 1).Unix(),
	}

	// Create token
	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(GetSecret()))
	if err != nil {
		return ""
	}
	return t
}

func NewAuthMiddleware(secret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(secret),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "wrong token",
			})
		},
	})
}

func Login(c *fiber.Ctx) error {
	// Extract the credentials from the request body
	log := new(loginRequest)
	// Parse the body into the loginRequest struct
	if err := c.BodyParser(log); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// Find the user by credentials
	if log.Email != "lucas@gmail.com" || log.Password != "123" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID := 1

	// Create the JWT claims, which includes the user ID and expiry time
	claims := jtoken.MapClaims{
		"ID":    userID,
		"exp":   time.Now().Add(time.Hour * 24 * 1).Unix(),
	}

	// Create token
	token := jtoken.NewWithClaims(jtoken.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return the token
	return c.JSON(fiber.Map{
		"token": t,
	})
}