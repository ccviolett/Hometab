package handler

import (
	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LinkGroupHandler struct {
	svc *service.LinkGroupSvc
}

func NewLinkGroupHandler(svc *service.LinkGroupSvc) *LinkGroupHandler {
	return &LinkGroupHandler{svc: svc}
}

func (h *LinkGroupHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.FindAll()
	if err != nil {
		return response.InternalError(c, "failed to fetch link groups")
	}
	return response.OK(c, items)
}

func (h *LinkGroupHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	item, err := h.svc.FindByID(id)
	if err != nil {
		return response.NotFound(c, "link group")
	}
	return response.OK(c, item)
}

func (h *LinkGroupHandler) Create(c *fiber.Ctx) error {
	var req model.LinkGroupCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Created(c, item)
}

func (h *LinkGroupHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.LinkGroupUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		return response.NotFound(c, "link group")
	}
	return response.OK(c, item)
}

func (h *LinkGroupHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	if err := h.svc.Delete(id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Message(c, "link group deleted")
}

func (h *LinkGroupHandler) Reorder(c *fiber.Ctx) error {
	var req model.ReorderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.Reorder(req.IDs); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Message(c, "link groups reordered")
}

func (h *LinkGroupHandler) ListByGroup(c *fiber.Ctx) error {
	result, err := h.svc.ListByGroup()
	if err != nil {
		return response.InternalError(c, "failed to fetch links by group")
	}
	return response.OK(c, result)
}
