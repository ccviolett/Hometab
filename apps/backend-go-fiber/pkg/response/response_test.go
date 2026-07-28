package response

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupApp() *fiber.App {
	app := fiber.New()
	app.Get("/ok", func(c *fiber.Ctx) error {
		return OK(c, fiber.Map{"key": "value"})
	})
	app.Get("/created", func(c *fiber.Ctx) error {
		return Created(c, fiber.Map{"id": 1})
	})
	app.Get("/message", func(c *fiber.Ctx) error {
		return Message(c, "done")
	})
	app.Get("/error", func(c *fiber.Ctx) error {
		return Error(c, 500, "something broke")
	})
	app.Get("/notfound", func(c *fiber.Ctx) error {
		return NotFound(c, "widget")
	})
	app.Get("/badrequest", func(c *fiber.Ctx) error {
		return BadRequest(c, "missing field")
	})
	app.Get("/internal", func(c *fiber.Ctx) error {
		return InternalError(c, "db down")
	})
	return app
}

func TestOK(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(body, &m)
	assert.Equal(t, "value", m["key"])
}

func TestCreated(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/created", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestMessage(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/message", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(body, &m)
	assert.Equal(t, "done", m["message"])
}

func TestErrorResponse(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(body, &m)
	assert.Equal(t, "something broke", m["detail"])
}

func TestNotFoundResponse(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/notfound", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(body, &m)
	assert.Equal(t, "widget not found", m["detail"])
}

func TestBadRequestResponse(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/badrequest", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestInternalErrorResponse(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/internal", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}
