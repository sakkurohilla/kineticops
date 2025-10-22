package handlers

import (
	"context"
	"strconv"

	"kineticops/backend/internal/services"
	"kineticops/backend/models"

	"github.com/gofiber/fiber/v2"
)

type HostHandler struct{ Service *services.HostService }

func (h *HostHandler) Register(r fiber.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Put("/:id", h.Update)
	r.Delete("/:id", h.Delete)
	r.Post("/:id/heartbeat", h.Heartbeat)
}

func (h *HostHandler) List(c *fiber.Ctx) error {
	ownerID := c.Locals("user_id").(int)
	data, err := h.Service.List(context.Background(), ownerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(data)
}

func (h *HostHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals("user_id").(int)
	var body models.Host
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	body.OwnerID = ownerID
	if err := h.Service.Create(context.Background(), &body); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(body)
}

func (h *HostHandler) Update(c *fiber.Ctx) error {
	owner := c.Locals("user_id").(int)
	id, _ := strconv.Atoi(c.Params("id"))
	var body models.Host
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.Service.Update(context.Background(), id, body.Name, body.IPAddress, owner); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *HostHandler) Delete(c *fiber.Ctx) error {
	owner := c.Locals("user_id").(int)
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Service.Delete(context.Background(), id, owner); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *HostHandler) Heartbeat(c *fiber.Ctx) error {
	owner := c.Locals("user_id").(int)
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Service.Heartbeat(context.Background(), id, owner); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusOK)
}
