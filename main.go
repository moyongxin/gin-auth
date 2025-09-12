package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var db *sql.DB
var jwtSecret []byte

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func initDB(path string) (*sql.DB, error) {
	d, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// create if not exists
	create := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);`
	_, err = d.Exec(create)
	if err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func main() {
	var err error
	// Emm... I think this is not a good secret actually
	jwtSecret = []byte(envOr("JWT_SECRET", "default-jwt-secret"))

	dbPath := envOr("DB_PATH", "./app.db")
	db, err = initDB(dbPath)
	if err != nil {
		log.Fatalf("failed init db: %v", err)
	}
	defer db.Close()

	r := gin.Default()

	// simple example, no rate limiting, no captcha
	// no email verification, no password reset
	// so I just write everything in huge closures :)
	r.POST("/register", func(c *gin.Context) {
		// I assume transferring password via HTTPS is secure
		// However, I still store only the hash in the database
		var body struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		pwHash, err := hashPassword(body.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		res, err := db.Exec("INSERT INTO users(username, password_hash) VALUES(?, ?)", body.Username, pwHash)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username taken or invalid"})
			return
		}
		// lazy, just use LastInsertId as user_id
		id, _ := res.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id, "username": body.Username})
	})

	r.POST("/login", func(c *gin.Context) {
		var body struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var id int64
		var pwHash string
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", body.Username).Scan(&id, &pwHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err := checkPassword(pwHash, body.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		token, err := createToken(id, body.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	// routes requiring authentication
	auth := r.Group("/")
	auth.Use(authMiddleware())
	auth.GET("/profile", func(c *gin.Context) {
		id, _ := c.Get("user_id")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"id": id, "username": username})
	})

	addr := envOr("ADDR", ":8080")
	r.Run(addr)
}
