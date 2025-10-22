package handlers

import (
	"database/sql"
	"kineticops/backend/config"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64  `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// RegisterUser handler
func RegisterUser(db *sqlx.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		var body request
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		if body.Email == "" || body.Password == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
		}

		// Check user exists
		var exists int
		err := db.Get(&exists, "SELECT count(*) FROM users WHERE email=$1", body.Email)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
		}
		if exists > 0 {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
		}

		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error processing password"})
		}

		// Insert user
		_, err = db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", body.Email, string(hashed))
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error creating user"})
		}

		return c.JSON(fiber.Map{"message": "user created successfully"})
	}
}

// LoginUser handler
func LoginUser(db *sqlx.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		var body request
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		if body.Email == "" || body.Password == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
		}

		var user User
		err := db.Get(&user, "SELECT id, email, password FROM users WHERE email=$1", body.Email)
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		claims := Claims{
			UserID: user.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte(cfg.JwtSecret))
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not create token"})
		}

		return c.JSON(fiber.Map{"token": signedToken})
	}
}

// ProtectedRoute handler
func ProtectedRoute(db *sqlx.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		const bearer = "Bearer "
		if len(auth) < len(bearer) || auth[:len(bearer)] != bearer {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header"})
		}

		tokenString := auth[len(bearer):]

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token claims"})
		}

		var user User
		err = db.Get(&user, "SELECT id, email FROM users WHERE id=$1", claims.UserID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "user not found"})
		}

		return c.JSON(fiber.Map{"id": user.ID, "email": user.Email})
	}
}
