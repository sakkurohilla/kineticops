package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

var syntheticService *services.SyntheticService

func InitSyntheticService() {
	syntheticService = services.NewSyntheticService()
}

// Monitor Management
func CreateSyntheticMonitor(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var monitor models.SyntheticMonitor
	if err := c.BodyParser(&monitor); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	monitor.TenantID = tid.(int64)

	if err := syntheticService.CreateMonitor(&monitor); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create monitor"})
	}

	return c.Status(201).JSON(monitor)
}

func GetSyntheticMonitors(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	monitors, err := syntheticService.GetMonitors(tid.(int64))
	if err != nil {
		return c.JSON([]models.SyntheticMonitor{})
	}

	return c.JSON(monitors)
}

func GetSyntheticMonitor(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	monitor, err := syntheticService.GetMonitor(id, tid.(int64))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Monitor not found"})
	}

	return c.JSON(monitor)
}

func UpdateSyntheticMonitor(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if err := syntheticService.UpdateMonitor(id, tid.(int64), updates); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot update monitor"})
	}

	return c.JSON(fiber.Map{"message": "Monitor updated"})
}

func DeleteSyntheticMonitor(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	
	if err := syntheticService.DeleteMonitor(id, tid.(int64)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot delete monitor"})
	}

	return c.JSON(fiber.Map{"message": "Monitor deleted"})
}

// Monitor Execution
func ExecuteSyntheticMonitor(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	monitor, err := syntheticService.GetMonitor(id, tid.(int64))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Monitor not found"})
	}

	result, err := syntheticService.ExecuteMonitor(monitor)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot execute monitor"})
	}

	return c.JSON(result)
}

// Results Management
func GetSyntheticResults(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	monitorID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	
	start, _ := time.Parse(time.RFC3339, c.Query("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
	end, _ := time.Parse(time.RFC3339, c.Query("end", time.Now().Format(time.RFC3339)))

	results, err := syntheticService.GetResults(monitorID, tid.(int64), start, end, limit)
	if err != nil {
		return c.JSON([]models.SyntheticResult{})
	}

	return c.JSON(results)
}

func GetSyntheticStats(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	monitorID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	
	start, _ := time.Parse(time.RFC3339, c.Query("start", time.Now().Add(-7*24*time.Hour).Format(time.RFC3339)))
	end, _ := time.Parse(time.RFC3339, c.Query("end", time.Now().Format(time.RFC3339)))

	stats, err := syntheticService.GetMonitorStats(monitorID, tid.(int64), start, end)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot fetch stats"})
	}

	return c.JSON(stats)
}

// Alert Management
func CreateSyntheticAlert(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var alert models.SyntheticAlert
	if err := c.BodyParser(&alert); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	alert.TenantID = tid.(int64)

	if err := syntheticService.CreateAlert(&alert); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot create alert"})
	}

	return c.Status(201).JSON(alert)
}

func GetSyntheticAlerts(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	monitorID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	
	alerts, err := syntheticService.GetAlerts(monitorID, tid.(int64))
	if err != nil {
		return c.JSON([]models.SyntheticAlert{})
	}

	return c.JSON(alerts)
}

// Browser Monitoring
func RecordBrowserSession(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var session models.BrowserSession
	if err := c.BodyParser(&session); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	session.TenantID = tid.(int64)
	if session.StartTime.IsZero() {
		session.StartTime = time.Now()
	}

	if err := syntheticService.RecordBrowserSession(&session); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot record session"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Session recorded"})
}

func RecordPageView(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var pageView models.PageView
	if err := c.BodyParser(&pageView); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	pageView.TenantID = tid.(int64)
	if pageView.Timestamp.IsZero() {
		pageView.Timestamp = time.Now()
	}

	if err := syntheticService.RecordPageView(&pageView); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot record page view"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Page view recorded"})
}

func RecordJavaScriptError(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var jsError models.JavaScriptError
	if err := c.BodyParser(&jsError); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	jsError.TenantID = tid.(int64)
	if jsError.Timestamp.IsZero() {
		jsError.Timestamp = time.Now()
	}

	if err := syntheticService.RecordJavaScriptError(&jsError); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot record error"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "JavaScript error recorded"})
}

func RecordAjaxRequest(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	var ajaxReq models.AjaxRequest
	if err := c.BodyParser(&ajaxReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	ajaxReq.TenantID = tid.(int64)
	if ajaxReq.Timestamp.IsZero() {
		ajaxReq.Timestamp = time.Now()
	}

	if err := syntheticService.RecordAjaxRequest(&ajaxReq); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot record AJAX request"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "AJAX request recorded"})
}

func GetBrowserSessions(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	appID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	
	start, _ := time.Parse(time.RFC3339, c.Query("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
	end, _ := time.Parse(time.RFC3339, c.Query("end", time.Now().Format(time.RFC3339)))

	sessions, err := syntheticService.GetBrowserSessions(appID, tid.(int64), start, end, limit)
	if err != nil {
		return c.JSON([]models.BrowserSession{})
	}

	return c.JSON(sessions)
}

func GetPageViews(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	appID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	
	start, _ := time.Parse(time.RFC3339, c.Query("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
	end, _ := time.Parse(time.RFC3339, c.Query("end", time.Now().Format(time.RFC3339)))

	pageViews, err := syntheticService.GetPageViews(appID, tid.(int64), start, end, limit)
	if err != nil {
		return c.JSON([]models.PageView{})
	}

	return c.JSON(pageViews)
}

func GetJavaScriptErrors(c *fiber.Ctx) error {
	tid := c.Locals("tenant_id")
	if tid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthenticated"})
	}

	appID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	
	start, _ := time.Parse(time.RFC3339, c.Query("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
	end, _ := time.Parse(time.RFC3339, c.Query("end", time.Now().Format(time.RFC3339)))

	errors, err := syntheticService.GetJavaScriptErrors(appID, tid.(int64), start, end, limit)
	if err != nil {
		return c.JSON([]models.JavaScriptError{})
	}

	return c.JSON(errors)
}