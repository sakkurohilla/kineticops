package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterSyntheticRoutes(app *fiber.App) {
	// Initialize Synthetic service
	handlers.InitSyntheticService()

	api := app.Group("/api/v1")
	synthetics := api.Group("/synthetics", middleware.AuthMiddleware)

	// Monitor Management
	synthetics.Post("/monitors", handlers.CreateSyntheticMonitor)
	synthetics.Get("/monitors", handlers.GetSyntheticMonitors)
	synthetics.Get("/monitors/:id", handlers.GetSyntheticMonitor)
	synthetics.Put("/monitors/:id", handlers.UpdateSyntheticMonitor)
	synthetics.Delete("/monitors/:id", handlers.DeleteSyntheticMonitor)

	// Monitor Execution
	synthetics.Post("/monitors/:id/execute", handlers.ExecuteSyntheticMonitor)

	// Results Management
	synthetics.Get("/monitors/:id/results", handlers.GetSyntheticResults)
	synthetics.Get("/monitors/:id/stats", handlers.GetSyntheticStats)

	// Alert Management
	synthetics.Post("/monitors/:id/alerts", handlers.CreateSyntheticAlert)
	synthetics.Get("/monitors/:id/alerts", handlers.GetSyntheticAlerts)

	// Browser Monitoring
	browser := api.Group("/browser", middleware.AuthMiddleware)
	browser.Post("/sessions", handlers.RecordBrowserSession)
	browser.Post("/pageviews", handlers.RecordPageView)
	browser.Post("/errors", handlers.RecordJavaScriptError)
	browser.Post("/ajax", handlers.RecordAjaxRequest)
	browser.Get("/applications/:id/sessions", handlers.GetBrowserSessions)
	browser.Get("/applications/:id/pageviews", handlers.GetPageViews)
	browser.Get("/applications/:id/errors", handlers.GetJavaScriptErrors)
}