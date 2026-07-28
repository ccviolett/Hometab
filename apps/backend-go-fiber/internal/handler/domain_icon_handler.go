package handler

import (
	"io"

	"hometab/internal/model"
	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
)

const maxUploadedIconBytes = 512 * 1024

type DomainIconHandler struct {
	svc     *service.DomainIconSvc
	linkSvc *service.LinkSvc
}

func NewDomainIconHandler(svc *service.DomainIconSvc, linkSvc *service.LinkSvc) *DomainIconHandler {
	return &DomainIconHandler{svc: svc, linkSvc: linkSvc}
}

func (h *DomainIconHandler) List(c *fiber.Ctx) error {
	items, err := h.svc.List()
	if err != nil {
		return response.InternalError(c, "failed to list icons")
	}
	return response.OK(c, items)
}

func (h *DomainIconHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}
	opened, err := file.Open()
	if err != nil {
		return response.BadRequest(c, "failed to open icon")
	}
	defer opened.Close()
	data, err := io.ReadAll(io.LimitReader(opened, maxUploadedIconBytes+1))
	if err != nil || len(data) > maxUploadedIconBytes {
		return response.BadRequest(c, "icon exceeds 512 KiB limit")
	}
	item, err := h.svc.Upload(c.Params("host"), data)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, item)
}

func (h *DomainIconHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.Params("host")); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Message(c, "icon deleted")
}

func (h *DomainIconHandler) Retry(c *fiber.Ctx) error {
	var req model.DomainIconCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	result, err := h.svc.Retry(c.Params("host"), req.URL)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, result)
}

func (h *DomainIconHandler) Resolve(c *fiber.Ctx) error {
	rawURL := c.Query("url")
	data, contentType, err := h.svc.Resolve(rawURL)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderCacheControl, "public, max-age=86400")
	return c.Send(data)
}

func (h *DomainIconHandler) Pending(c *fiber.Ctx) error {
	host := c.Params("host")
	data, contentType, err := h.svc.Pending(host)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.Send(data)
}

func (h *DomainIconHandler) Check(c *fiber.Ctx) error {
	var req model.DomainIconCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	result, err := h.svc.Check(req.URL)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, result)
}

func (h *DomainIconHandler) Choose(c *fiber.Ctx) error {
	host := c.Params("host")
	var req model.DomainIconChooseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	item, err := h.svc.Choose(host, req.Choice)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, item)
}

func (h *DomainIconHandler) RefreshAll(c *fiber.Ctx) error {
	links, err := h.linkSvc.FindAll()
	if err != nil {
		return response.InternalError(c, "failed to fetch links")
	}
	result := model.DomainIconRefreshAllResponse{
		TotalLinks: len(links),
		Errors:     make([]string, 0),
	}
	seen := make(map[string]string)
	for _, link := range links {
		host, err := service.HostForURL(link.URL)
		if err != nil || host == "" {
			result.Failed++
			if len(result.Errors) < 10 {
				result.Errors = append(result.Errors, link.URL+": invalid url")
			}
			continue
		}
		if _, ok := seen[host]; !ok {
			seen[host] = link.URL
		}
	}
	result.TotalHosts = len(seen)
	for host, rawURL := range seen {
		check, err := h.svc.Check(rawURL)
		if err != nil {
			result.Failed++
			if len(result.Errors) < 10 {
				result.Errors = append(result.Errors, host+": "+err.Error())
			}
			continue
		}
		switch check.Status {
		case "ready":
			result.Ready++
		case "unchanged":
			result.Unchanged++
		case "conflict":
			result.Conflicts++
		default:
			result.Failed++
			if check.Error != "" && len(result.Errors) < 10 {
				result.Errors = append(result.Errors, host+": "+check.Error)
			}
		}
	}
	if len(result.Errors) == 0 {
		result.Errors = nil
	}
	return response.OK(c, result)
}
