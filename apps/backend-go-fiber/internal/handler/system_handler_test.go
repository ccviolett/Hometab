package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http/httptest"
	"testing"

	"hometab/internal/autostart"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStartupManager struct {
	status    autostart.Status
	configure autostart.Config
	err       error
}

func (m *fakeStartupManager) Status() (autostart.Status, error) {
	return m.status, m.err
}

func (m *fakeStartupManager) Configure(config autostart.Config) (autostart.Status, error) {
	m.configure = config
	return m.status, m.err
}

func TestSystemStartupStatusAndConfigure(t *testing.T) {
	manager := &fakeStartupManager{status: autostart.Status{
		Platform: "darwin", Supported: true, Enabled: true,
	}}
	app := fiber.New()
	handler := testSystemHandler(manager, true)
	app.Get("/api/system/startup", handler.StartupStatus)
	app.Put("/api/system/startup", handler.ConfigureStartup)

	request := httptest.NewRequest("GET", "/api/system/startup", nil)
	response, err := app.Test(request)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)

	body, err := json.Marshal(autostart.Config{Enabled: true})
	require.NoError(t, err)
	request = httptest.NewRequest("PUT", "/api/system/startup", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err = app.Test(request)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	assert.Equal(t, autostart.Config{Enabled: true}, manager.configure)
}

func TestSystemStartupRejectsRemoteClients(t *testing.T) {
	app := fiber.New()
	handler := testSystemHandler(&fakeStartupManager{}, false)
	app.Get("/api/system/startup", handler.StartupStatus)

	request := httptest.NewRequest("GET", "/api/system/startup", nil)
	response, err := app.Test(request)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, response.StatusCode)
}

func TestSystemStartupReportsUnsupportedPlatform(t *testing.T) {
	app := fiber.New()
	handler := testSystemHandler(&fakeStartupManager{err: autostart.ErrUnsupported}, true)
	app.Put("/api/system/startup", handler.ConfigureStartup)

	request := httptest.NewRequest("PUT", "/api/system/startup", bytes.NewBufferString(`{"enabled":true}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, response.StatusCode)
}

func TestSystemStartupReturnsManagerErrors(t *testing.T) {
	app := fiber.New()
	handler := testSystemHandler(&fakeStartupManager{err: errors.New("boom")}, true)
	app.Get("/api/system/startup", handler.StartupStatus)

	request := httptest.NewRequest("GET", "/api/system/startup", nil)
	response, err := app.Test(request)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
}

func TestSystemRequestAllowed(t *testing.T) {
	tests := []struct {
		name     string
		remoteIP string
		hostname string
		origin   string
		allowed  bool
	}{
		{name: "local IPv4", remoteIP: "127.0.0.1", hostname: "127.0.0.1:52173", allowed: true},
		{name: "local hostname and origin", remoteIP: "::1", hostname: "localhost:52173", origin: "http://localhost:52173", allowed: true},
		{name: "local IPv6", remoteIP: "::1", hostname: "[::1]:52173", allowed: true},
		{name: "remote client", remoteIP: "192.0.2.10", hostname: "127.0.0.1", allowed: false},
		{name: "proxied public host", remoteIP: "127.0.0.1", hostname: "hometab.example.com", allowed: false},
		{name: "foreign browser origin", remoteIP: "127.0.0.1", hostname: "127.0.0.1", origin: "https://example.com", allowed: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.allowed, systemRequestAllowed(net.ParseIP(test.remoteIP), test.hostname, test.origin))
		})
	}
}

func testSystemHandler(manager startupManager, allowed bool) *SystemHandler {
	return &SystemHandler{
		startup: manager,
		allowed: func(*fiber.Ctx) bool { return allowed },
	}
}
