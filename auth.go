package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// thank bcrypt for password hashing,
// I'm not sure about using hash for storing passwords
// but at least don't store plain text passwords
func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// only use sub, usr, exp claims
// omitting iss, aud, iat, nbf claims
// because I'm too lazy to populate too many fields for this simple example
func createToken(userID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"usr": username,
		"exp": time.Now().Add(24 * time.Hour).Unix(), // 24 hours expiration
	}
	// I'm not sure about signing methods in JWT, use HMAC+SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// auth => "Bearer <token>"
		auth := c.GetHeader("Authorization")
		if len(auth) < 7 || auth[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tok := auth[7:]
		parsed, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) {
			// HS256 should be a member of HMAC family
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			// passed
			return jwtSecret, nil
		})
		if err != nil || !parsed.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "detail": err.Error()})
			return
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}
		// sub can be number or string
		var uid int64
		switch v := claims["sub"].(type) {
		// json quirk makes all numbers float64
		case float64:
			uid = int64(v)
		case string:
			fmt.Sscan(v, &uid)
		}
		username, _ := claims["usr"].(string)
		c.Set("user_id", uid)
		c.Set("username", username)
		c.Next()
	}
}
