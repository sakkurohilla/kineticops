package main

import (
	"kineticops/backend/config"
	"kineticops/backend/handlers"
	"kineticops/backend/internal/repo"
	"kineticops/backend/internal/routes"
	"kineticops/backend/internal/services"
	"kineticops/backend/middleware"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("DB connect error: %v\n", err)
	}
	defer db.Close()

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	hostRepo := &repo.HostRepo{DB: db}
	hostSvc := &services.HostService{Repo: hostRepo}
	hostHandler := &handlers.HostHandler{Service: hostSvc}
	routes.RegisterHostRoutes(app, hostHandler)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://127.0.0.1:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Post("/api/register", handlers.RegisterUser(db, cfg))
	app.Post("/api/login", handlers.LoginUser(db, cfg))
	app.Get("/api/profile", middleware.Protected(db, cfg), handlers.ProtectedRoute(db, cfg))

	app.Get("/api/workspaces", middleware.Protected(db, cfg), handlers.ListWorkspaces(db))
	app.Post("/api/workspaces", middleware.Protected(db, cfg), handlers.CreateWorkspace(db))
	app.Put("/api/workspaces/:id", middleware.Protected(db, cfg), handlers.UpdateWorkspace(db))
	app.Delete("/api/workspaces/:id", middleware.Protected(db, cfg), handlers.DeleteWorkspace(db))

	log.Fatal(app.Listen(cfg.ListenAddress))
}
