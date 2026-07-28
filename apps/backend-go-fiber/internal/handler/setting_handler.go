package handler

import (
	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type SettingHandler struct {
	svc *service.SettingSvc
}

func NewSettingHandler(svc *service.SettingSvc) *SettingHandler {
	return &SettingHandler{svc: svc}
}

func (h *SettingHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.FindAll()
	if err != nil {
		return response.InternalError(c, "failed to fetch settings: "+err.Error())
	}
	return response.OK(c, items)
}

func (h *SettingHandler) Get(c *fiber.Ctx) error {
	key := c.Params("key")
	item, err := h.svc.FindByKey(key)
	if err != nil {
		return response.NotFound(c, "setting")
	}
	return response.OK(c, item)
}

func (h *SettingHandler) CreateOrUpdate(c *fiber.Ctx) error {
	var req model.SettingCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Key == "" {
		return response.BadRequest(c, "key is required")
	}
	item, err := h.svc.CreateOrUpdate(req)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, item)
}

func (h *SettingHandler) Update(c *fiber.Ctx) error {
	key := c.Params("key")
	var req model.SettingUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Update(key, req)
	if err != nil {
		return response.NotFound(c, "setting")
	}
	return response.OK(c, item)
}

func (h *SettingHandler) Delete(c *fiber.Ctx) error {
	key := c.Params("key")
	if err := h.svc.Delete(key); err != nil {
		return response.InternalError(c, "failed to delete setting")
	}
	return response.Message(c, "setting deleted")
}
