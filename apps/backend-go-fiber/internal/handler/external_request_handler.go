package handler

import (
	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ExternalRequestHandler struct {
	svc *service.ExternalRequestSvc
}

func NewExternalRequestHandler(svc *service.ExternalRequestSvc) *ExternalRequestHandler {
	return &ExternalRequestHandler{svc: svc}
}

func (h *ExternalRequestHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.FindAll()
	if err != nil {
		return response.InternalError(c, "failed to fetch external requests")
	}
	return response.OK(c, items)
}

func (h *ExternalRequestHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	item, err := h.svc.FindByID(id)
	if err != nil {
		return response.NotFound(c, "external request")
	}
	return response.OK(c, item)
}

func (h *ExternalRequestHandler) Create(c *fiber.Ctx) error {
	var req model.ExternalRequestCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, item)
}

func (h *ExternalRequestHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.ExternalRequestUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, item)
}

func (h *ExternalRequestHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	if err := h.svc.Delete(id); err != nil {
		return response.InternalError(c, "failed to delete external request")
	}
	return response.Message(c, "external request deleted")
}

func (h *ExternalRequestHandler) Execute(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	result, err := h.svc.Execute(id)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, result)
}
