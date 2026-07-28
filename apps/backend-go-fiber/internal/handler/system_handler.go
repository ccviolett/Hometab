package handler

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"hometab/internal/autostart"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type startupManager interface {
	Status() (autostart.Status, error)
	Configure(autostart.Config) (autostart.Status, error)
}

type SystemHandler struct {
	startup startupManager
	allowed func(*fiber.Ctx) bool
}

func NewSystemHandler(startup startupManager) *SystemHandler {
	return &SystemHandler{startup: startup, allowed: isLoopbackRequest}
}

func (h *SystemHandler) StartupStatus(c *fiber.Ctx) error {
	if !h.allowed(c) {
		return response.Error(c, fiber.StatusForbidden, "system settings are available only from this device")
	}
	status, err := h.startup.Status()
	if err != nil {
		return response.InternalError(c, "failed to read login startup status: "+err.Error())
	}
	return response.OK(c, status)
}

func (h *SystemHandler) ConfigureStartup(c *fiber.Ctx) error {
	if !h.allowed(c) {
		return response.Error(c, fiber.StatusForbidden, "system settings are available only from this device")
	}
	var config autostart.Config
	if err := c.BodyParser(&config); err != nil {
		return response.BadRequest(c, "invalid login startup configuration")
	}
	status, err := h.startup.Configure(config)
	if errors.Is(err, autostart.ErrUnsupported) {
		return response.Error(c, fiber.StatusNotImplemented, err.Error())
	}
	if err != nil {
		return response.InternalError(c, "failed to configure login startup: "+err.Error())
	}
	return response.OK(c, status)
}

func isLoopbackRequest(c *fiber.Ctx) bool {
	return systemRequestAllowed(c.Context().RemoteIP(), c.Hostname(), c.Get(fiber.HeaderOrigin))
}

func systemRequestAllowed(remoteIP net.IP, hostname, origin string) bool {
	if !remoteIP.IsLoopback() || !isLoopbackHostname(hostname) {
		return false
	}
	if strings.TrimSpace(origin) == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	return err == nil && isLoopbackHostname(parsed.Hostname())
}

func isLoopbackHostname(hostname string) bool {
	hostname = strings.TrimSpace(hostname)
	if host, _, err := net.SplitHostPort(hostname); err == nil {
		hostname = host
	}
	hostname = strings.Trim(hostname, "[]")
	if strings.EqualFold(hostname, "localhost") {
		return true
	}
	ip := net.ParseIP(hostname)
	return ip != nil && ip.IsLoopback()
}
