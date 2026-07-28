package handler

import (
	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LinkFlowHandler struct {
	svc *service.LinkFlowSvc
}

func NewLinkFlowHandler(svc *service.LinkFlowSvc) *LinkFlowHandler {
	return &LinkFlowHandler{svc: svc}
}

func (h *LinkFlowHandler) List(c *fiber.Ctx) error {
	groupID := c.Query("group_id")
	var gidPtr *string
	if groupID != "" {
		gidPtr = &groupID
	}
	items, err := h.svc.FindAll(gidPtr)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, items)
}

func (h *LinkFlowHandler) Create(c *fiber.Ctx) error {
	var req model.LinkFlowCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" || req.GroupID == "" {
		return response.BadRequest(c, "name and group_id are required")
	}
	item, err := h.svc.Create(req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, item)
}

func (h *LinkFlowHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var req model.LinkFlowUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Update(id, req)
	if err != nil {
		return response.NotFound(c, "link flow")
	}
	return response.OK(c, item)
}

func (h *LinkFlowHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}
	var opts model.LinkFlowDeleteOptions
	_ = c.BodyParser(&opts)
	if err := h.svc.Delete(id, &opts); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Message(c, "link flow deleted")
}

func (h *LinkFlowHandler) AddLink(c *fiber.Ctx) error {
	flowID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid flow id")
	}
	var req model.LinkFlowLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.LinkID == "" {
		return response.BadRequest(c, "link_id is required")
	}
	link, err := h.svc.AddLink(flowID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, link)
}

func (h *LinkFlowHandler) UpdateLink(c *fiber.Ctx) error {
	flowID, err := uuid.Parse(c.Params("fid"))
	if err != nil {
		return response.BadRequest(c, "invalid flow id")
	}
	linkID, err := uuid.Parse(c.Params("lid"))
	if err != nil {
		return response.BadRequest(c, "invalid link id")
	}
	var req model.LinkFlowLinkOrderUpdate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	link, err := h.svc.UpdateLinkOrder(flowID, linkID, req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, link)
}

func (h *LinkFlowHandler) ReorderLinks(c *fiber.Ctx) error {
	flowID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid flow id")
	}
	var req model.ReorderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.ReorderLinks(flowID, req.IDs); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Message(c, "flow links reordered")
}

func (h *LinkFlowHandler) RemoveLink(c *fiber.Ctx) error {
	flowID, err := uuid.Parse(c.Params("fid"))
	if err != nil {
		return response.BadRequest(c, "invalid flow id")
	}
	linkID, err := uuid.Parse(c.Params("lid"))
	if err != nil {
		return response.BadRequest(c, "invalid link id")
	}
	if err := h.svc.RemoveLink(flowID, linkID); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Message(c, "link removed from flow")
}
