package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/kineticops/backend/internal/auth"
	"github.com/kineticops/backend/internal/models"
	"github.com/kineticops/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserRepo   *repository.UserRepository
	JWTService *auth.JWTService
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{UserRepo: userRepo, JWTService: jwtService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type request struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var body request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	user := models.NewUser(body.Name, body.Email, string(hashed))
	if err := h.UserRepo.CreateUser(context.Background(), user); err != nil {
		log.Println("CreateUser error:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "cannot create user"})
	}
	token, _ := h.JWTService.GenerateToken(user.ID.String())
	return c.JSON(fiber.Map{"token": token})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var body request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}
	user, err := h.UserRepo.GetUserByEmail(context.Background(), body.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}
	token, _ := h.JWTService.GenerateToken(user.ID.String())
	return c.JSON(fiber.Map{"token": token})
}
