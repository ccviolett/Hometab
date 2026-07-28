package handler

import (
	"strconv"

	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type SearchEngineHandler struct {
	svc *service.SearchEngineSvc
}

func NewSearchEngineHandler(svc *service.SearchEngineSvc) *SearchEngineHandler {
	return &SearchEngineHandler{svc: svc}
}

func (h *SearchEngineHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.FindAll()
	if err != nil {
		return response.InternalError(c, "failed to fetch search engines")
	}
	return response.OK(c, items)
}

func (h *SearchEngineHandler) Create(c *fiber.Ctx) error {
	var req model.SearchEngineCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" || req.URLTemplate == "" {
		return response.BadRequest(c, "name and url_template are required")
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Created(c, item)
}

func (h *SearchEngineHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.SearchEngineUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Update(uint(id), req)
	if err != nil {
		return response.NotFound(c, "search engine")
	}
	return response.OK(c, item)
}

func (h *SearchEngineHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		return response.InternalError(c, "failed to delete search engine")
	}
	return response.Message(c, "search engine deleted")
}
