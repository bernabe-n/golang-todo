package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization") //Reads the Authorization header from the request

		if authHeader == "" { //Check if header is missing
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"}) //Respond with 401 Unauthorized
			c.Abort()                                                                        //stops the request from continuing
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ") //Remove "Bearer " prefix, Ex: "Bearer abc123" → "abc123"

		if tokenString == "" || tokenString == authHeader { //Validate token format, if empty token or "Bearer " was not present
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { //Parses and validates the token, Takes: token string and a callback function that returns the secret key
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() { //Ensures the token uses HS256 algorithm
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil //Converts your secret into []byte, Used to verify the token signature
		})

		if err != nil || !token.Valid { //parsing failed OR token is invalid
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims) //Extract claims, JWT payload = claims

		if !ok { //Validate claims type
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token Claims"}) //If casting failed → invalid token
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string) //Gets user_id from token and Casts it to string

		if !ok { //Validate user_id
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"}) //If missing or wrong type
			c.Abort()
			return
		}

		if exp, ok := claims["exp"].(float64); ok { //Check expiration manually, exp = expiration timestamp, JWT stores it as float64
			expirationTime := time.Unix(int64(exp), 0) //Converts timestamp → time.Time

			if time.Now().After(expirationTime) { //Checks if current time is past expiration
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				c.Abort()
				return
			}
		}

		c.Set("user_id", userID) //Saves user ID in request context, You can access it later in handlers:
		c.Next()                 //Allows request to continue to the actual route handler
	}
}

/*
Big Picture (what this middleware does)
Gets JWT from request header
Validates format (Bearer token)
Parses and verifies signature
Checks claims (user_id + expiration)
Stores user_id in context
Allows request if valid, blocks if not
*/
