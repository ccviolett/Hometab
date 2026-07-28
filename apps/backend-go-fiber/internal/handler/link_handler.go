package handler

import (
	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LinkHandler struct {
	svc *service.LinkSvc
}

func NewLinkHandler(svc *service.LinkSvc) *LinkHandler {
	return &LinkHandler{svc: svc}
}

func (h *LinkHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.FindAll()
	if err != nil {
		return response.InternalError(c, "failed to fetch links")
	}
	return response.OK(c, items)
}

func (h *LinkHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	item, err := h.svc.FindByID(id)
	if err != nil {
		return response.NotFound(c, "link")
	}
	return response.OK(c, item)
}

func (h *LinkHandler) Create(c *fiber.Ctx) error {
	var req model.LinkCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" || req.URL == "" {
		return response.BadRequest(c, "name and url are required")
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, item)
}

func (h *LinkHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.LinkUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, item)
}

func (h *LinkHandler) ReorderGroup(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("groupID"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}
	var req model.ReorderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.ReorderGroup(groupID, req.IDs); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Message(c, "links reordered")
}

func (h *LinkHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	if err := h.svc.Delete(id); err != nil {
		return response.InternalError(c, "failed to delete link")
	}
	return response.Message(c, "link deleted")
}
