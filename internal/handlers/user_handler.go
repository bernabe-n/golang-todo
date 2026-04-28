package handlers

import (
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/models"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct { //Defines a struct (like a data container) for incoming JSON
	Email    string `json:"email" binding:"required"` //maps JSON field "email" → Go field Email, Gin validates that this field must exist in the request
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct { //Defines the shape of incoming login data
	Email    string `json:"email" binding:"required"` //When reading JSON, map "email" → Email
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct { //Defines the shape of the response you send back after login, Usually contains a JWT or session token
	Token string `json:"token"`
}

func CreateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc { //This is called a closure pattern — you pass the DB into the handler
	return func(c *gin.Context) { //represents the HTTP request + response
		var registerRequest RegisterRequest //Creates an empty struct to store incoming JSON data

		if err := c.BindJSON(&registerRequest); err != nil { //Reads JSON body from request, Converts it into the struct
			c.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}

		if len(registerRequest.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long!"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost) //Converts password string → []byte, Hashes it using bcrypt, DefaultCost controls how strong (slow) the hashing i

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password" + err.Error()}) // If hashing fails
			return
		}

		user := &models.User{ // Creates a new User struct
			Email:    registerRequest.Email,  //Email from request
			Password: string(hashedPassword), //Hashed password (converted back to string)
		}

		createdUser, err := repository.CreateUser(pool, user) //Calls your repository function, Returns: createdUser (saved user from DB)

		if err != nil { //Handle database errors
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") { //Checks if error is about: duplicate email, unique constraint violation
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, createdUser)
	}
}

func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc { //cfg → config (contains things like your JWT secret)
	return func(c *gin.Context) {
		var loginRequest LoginRequest //Creates an empty struct, Will hold: email and password

		if err := c.BindJSON(&loginRequest); err != nil { //Reads JSON body from client, Converts it into loginRequest
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) //If invalid JSON → return 400 Bad Request
			return
		}

		user, err := repository.GetUserByEmail(pool, loginRequest.Email) //Looks up user by email in database
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials"}) //If user not found → return 401 Unauthorized
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)) //Compares:stored hashed password (DB) & plain password (user input)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials"}) //If passwords don’t match → 401 Unauthorized
			return
		}

		//map[string]any{}
		claims := jwt.MapClaims{ //a map (like map[string]interface{}) and stores in token
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(), //Token becomes invalid after 24 hours
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //Creates a JWT object, Uses: Algorithm: HS256 (HMAC + SHA256), Claims: the data you just defined

		tokenString, err := token.SignedString([]byte(cfg.JWTSecret)) //Signs the token using your secret key, cfg.JWTSecret = secret from config
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()}) //If signing fails → return 500 error
			return
		}

		c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
	}
}

func TestProtectedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")

		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Protected route accessed successfully",
			"user_id": userID,
		})
	}
}

/*
Big picture flow in CreateUserHandler
Receive HTTP request
Parse JSON → struct
Validate password
Hash password
Create user object
Insert into DB
Handle errors (duplicate, etc.)
Return success response
*/

/*Full flow (big picture) in LoginHandler
Receive login request (email + password)
Parse JSON
Check if user exists
Compare password with hashed password
Create JWT payload (claims)
Sign token with secret
Return token to client
*/
