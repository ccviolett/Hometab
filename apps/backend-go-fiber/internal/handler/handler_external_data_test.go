package handler_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalRequestsCRUDAndExecute(t *testing.T) {
	app := newTestApp(t)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token-1", r.Header.Get("X-Test"))
		assert.Equal(t, "home", r.URL.Query().Get("source"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"total":7},"message":"ok"}`))
	}))
	defer target.Close()

	resp, created := doJSON(app, "POST", "/api/external-requests", map[string]interface{}{
		"name":               "Service Count",
		"method":             "GET",
		"url":                target.URL,
		"headers_json":       `{"X-Test":"token-1"}`,
		"query_json":         `{"source":"home"}`,
		"parser_type":        "json_path",
		"parser_config_json": `[{"label":"服务数量","path":"$.data.total"},{"label":"消息","path":"$.message"}]`,
	})
	require.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "Service Count", created["name"])
	assert.Equal(t, "GET", created["method"])
	assert.Equal(t, false, created["confirm_before_run"])
	id := created["id"].(string)

	resp, got := doJSON(app, "GET", "/api/external-requests/"+id, nil)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, target.URL, got["url"])

	resp, updated := doJSON(app, "PUT", "/api/external-requests/"+id, map[string]interface{}{
		"name": "Service Count Updated",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Service Count Updated", updated["name"])

	resp, result := doJSON(app, "POST", "/api/external-requests/"+id+"/execute", nil)
	require.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, float64(200), result["status"])
	parsed := result["parsed"].([]interface{})
	require.Len(t, parsed, 2)
	first := parsed[0].(map[string]interface{})
	assert.Equal(t, "服务数量", first["label"])
	assert.Equal(t, float64(7), first["value"])

	resp, arr := doJSONArray(app, "GET", "/api/external-requests")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, arr, 1)

	resp, _ = doJSON(app, "DELETE", "/api/external-requests/"+id, nil)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestExternalRequestsValidation(t *testing.T) {
	app := newTestApp(t)

	resp, _ := doJSON(app, "POST", "/api/external-requests", map[string]interface{}{
		"name": "Bad",
		"url":  "ftp://example.com",
	})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "POST", "/api/external-requests", map[string]interface{}{
		"name":   "Bad Method",
		"url":    "https://example.com",
		"method": "TRACE",
	})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "POST", "/api/external-requests", map[string]interface{}{
		"name":         "Bad Headers",
		"url":          "https://example.com",
		"headers_json": `[]`,
	})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "POST", "/api/external-requests", map[string]interface{}{
		"name":   "Delete Action",
		"url":    "https://example.com",
		"method": "DELETE",
	})
	assert.Equal(t, 201, resp.StatusCode)
}

// ==================== Settings CRUD ====================

func TestSettingsCRUD(t *testing.T) {
	app := newTestApp(t)

	// CreateOrUpdate
	resp, created := doJSON(app, "POST", "/api/settings", map[string]interface{}{
		"key":   "theme",
		"value": "dark",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "theme", created["key"])
	assert.Equal(t, "dark", created["value"])

	// Get by key
	resp, got := doJSON(app, "GET", "/api/settings/theme", nil)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "dark", got["value"])

	// Update
	resp, updated := doJSON(app, "PUT", "/api/settings/theme", map[string]interface{}{
		"value": "light",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "light", updated["value"])

	// Upsert (same key again)
	resp, _ = doJSON(app, "POST", "/api/settings", map[string]interface{}{
		"key":   "theme",
		"value": "auto",
	})
	assert.Equal(t, 200, resp.StatusCode)

	// List (returns map)
	req := httptest.NewRequest("GET", "/api/settings", nil)
	listResp, _ := app.Test(req, -1)
	assert.Equal(t, 200, listResp.StatusCode)

	// Delete
	resp, _ = doJSON(app, "DELETE", "/api/settings/theme", nil)
	assert.Equal(t, 200, resp.StatusCode)

	// Not found after delete
	resp, _ = doJSON(app, "GET", "/api/settings/theme", nil)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestSettingsValidation(t *testing.T) {
	app := newTestApp(t)

	// Missing key
	resp, _ := doJSON(app, "POST", "/api/settings", map[string]interface{}{
		"value": "x",
	})
	assert.Equal(t, 400, resp.StatusCode)

	// Update non-existent
	resp, _ = doJSON(app, "PUT", "/api/settings/nonexistent", map[string]interface{}{
		"value": "x",
	})
	assert.Equal(t, 404, resp.StatusCode)
}

// ==================== Data Export/Import ====================

func TestDataExportImport(t *testing.T) {
	app := newTestApp(t)

	// Seed some data
	doJSON(app, "POST", "/api/settings", map[string]interface{}{
		"key": "theme", "value": "dark",
	})
	doJSON(app, "POST", "/api/search-engines", map[string]interface{}{
		"name": "Google", "url_template": "https://google.com/search?q={query}",
	})
	doJSON(app, "POST", "/api/external-requests", map[string]interface{}{
		"name": "Ping", "url": "https://example.com", "method": "GET",
	})

	// Export
	req := httptest.NewRequest("GET", "/api/export", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "zip")

	zipData, _ := io.ReadAll(resp.Body)
	assert.Greater(t, len(zipData), 100)

	// Import into a fresh app via multipart form
	app2 := newTestApp(t)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "backup.zip")
	require.NoError(t, err)
	_, err = part.Write(zipData)
	require.NoError(t, err)
	writer.Close()

	importReq := httptest.NewRequest("POST", "/api/import", &buf)
	importReq.Header.Set("Content-Type", writer.FormDataContentType())
	importResp, err := app2.Test(importReq, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, importResp.StatusCode)

	// Verify imported data
	resp2, setting := doJSON(app2, "GET", "/api/settings/theme", nil)
	assert.Equal(t, 200, resp2.StatusCode)
	assert.Equal(t, "dark", setting["value"])

	resp2, requests := doJSONArray(app2, "GET", "/api/external-requests")
	assert.Equal(t, 200, resp2.StatusCode)
	assert.Len(t, requests, 1)
}

func TestUnsupportedEndpointsReturnNotFound(t *testing.T) {
	app := newTestApp(t)

	removed := []string{
		"/api/services",
		"/api/icon",
		"/api/plugins",
		"/api/servicehub/health",
		"/api/servicehub/running-services",
		"/api/newsfeed/github/trending",
		"/api/newsfeed/juejin/articles",
		"/api/newsfeed/zhihu/recommend",
		"/api/newsfeed/rebang/items",
	}
	for _, ep := range removed {
		req := httptest.NewRequest("GET", ep, nil)
		resp, _ := app.Test(req, -1)
		assert.Equal(t, 404, resp.StatusCode, "endpoint: %s", ep)
	}
}
