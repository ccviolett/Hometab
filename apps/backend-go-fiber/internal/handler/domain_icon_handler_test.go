package handler_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainIconResolveFallback(t *testing.T) {
	app := newTestApp(t)
	req := httptest.NewRequest("GET", "/api/link-icons/resolve?url=https%3A%2F%2Fexample.invalid", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "image/svg+xml", resp.Header.Get("Content-Type"))
}

func TestDomainIconCheckValidation(t *testing.T) {
	app := newTestApp(t)
	resp, m := doJSON(app, "POST", "/api/link-icons/check", map[string]interface{}{"url": ""})
	assert.Equal(t, 400, resp.StatusCode)
	assert.NotEmpty(t, m["detail"])
}
