package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"hometab/internal/testutil"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	app, _ := testutil.SetupTestApp()
	return app
}

func doJSON(app *fiber.App, method, url string, body interface{}) (*http.Response, map[string]interface{}) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	data, _ := io.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return resp, m
}

func doJSONArray(app *fiber.App, method, url string) (*http.Response, []interface{}) {
	req := httptest.NewRequest(method, url, nil)
	resp, _ := app.Test(req, -1)
	data, _ := io.ReadAll(resp.Body)
	var arr []interface{}
	json.Unmarshal(data, &arr)
	return resp, arr
}

// ==================== Health & Build Info ====================

func TestHealth(t *testing.T) {
	app := newTestApp(t)
	resp, m := doJSON(app, "GET", "/health", nil)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "healthy", m["status"])
}

func TestBuildInfo(t *testing.T) {
	app := newTestApp(t)
	resp, m := doJSON(app, "GET", "/api/build-info", nil)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "test", m["version"])
	assert.Equal(t, "2025-01-01 00:00:00 UTC", m["build_time"])
	assert.NotEmpty(t, m["go_version"])
}

// ==================== Links CRUD ====================

func TestLinksCRUD(t *testing.T) {
	app := newTestApp(t)

	// Create a group first
	resp, group := doJSON(app, "POST", "/api/link-groups", map[string]interface{}{
		"name": "TestGroup",
	})
	require.Equal(t, 201, resp.StatusCode)
	groupID := group["id"].(string)

	// Create link
	resp, created := doJSON(app, "POST", "/api/links", map[string]interface{}{
		"name":     "Google",
		"url":      "https://google.com",
		"group_id": groupID,
	})
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "Google", created["name"])
	linkID := created["id"].(string)

	// Get
	resp, got := doJSON(app, "GET", "/api/links/"+linkID, nil)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Google", got["name"])

	// Update
	resp, updated := doJSON(app, "PUT", "/api/links/"+linkID, map[string]interface{}{
		"name": "Bing",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Bing", updated["name"])

	// List
	resp, arr := doJSONArray(app, "GET", "/api/links")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, arr, 1)

	// Delete
	resp, _ = doJSON(app, "DELETE", "/api/links/"+linkID, nil)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLinksValidation(t *testing.T) {
	app := newTestApp(t)

	resp, _ := doJSON(app, "POST", "/api/links", map[string]interface{}{
		"name": "NoURL",
	})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "POST", "/api/links", map[string]interface{}{
		"name":     "BadGroup",
		"url":      "https://x.com",
		"group_id": "not-a-uuid",
	})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "GET", "/api/links/invalid", nil)
	assert.Equal(t, 400, resp.StatusCode)
}

// ==================== Link Groups CRUD ====================

func TestLinkGroupsCRUD(t *testing.T) {
	app := newTestApp(t)

	// Create
	desc := "A test group"
	resp, created := doJSON(app, "POST", "/api/link-groups", map[string]interface{}{
		"name":        "MyGroup",
		"description": desc,
		"order_index": 1,
	})
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "MyGroup", created["name"])
	id := created["id"].(string)

	// Get
	resp, got := doJSON(app, "GET", "/api/link-groups/"+id, nil)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "MyGroup", got["name"])

	// Update
	resp, updated := doJSON(app, "PUT", "/api/link-groups/"+id, map[string]interface{}{
		"name": "Renamed",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Renamed", updated["name"])

	// List
	resp, arr := doJSONArray(app, "GET", "/api/link-groups")
	assert.Equal(t, 200, resp.StatusCode)
	assert.GreaterOrEqual(t, len(arr), 1)

	// Delete (should create "未分组" group and move links)
	resp, _ = doJSON(app, "DELETE", "/api/link-groups/"+id, nil)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLinkGroupsValidation(t *testing.T) {
	app := newTestApp(t)

	resp, _ := doJSON(app, "POST", "/api/link-groups", map[string]interface{}{})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "GET", "/api/link-groups/bad-id", nil)
	assert.Equal(t, 400, resp.StatusCode)
}

// ==================== Links By Group ====================

func TestLinksByGroup(t *testing.T) {
	app := newTestApp(t)

	// Create group
	resp, group := doJSON(app, "POST", "/api/link-groups", map[string]interface{}{
		"name": "G1",
	})
	require.Equal(t, 201, resp.StatusCode)
	gid := group["id"].(string)

	// Create link in group
	resp, _ = doJSON(app, "POST", "/api/links", map[string]interface{}{
		"name":     "Link1",
		"url":      "https://a.com",
		"group_id": gid,
	})
	require.Equal(t, 201, resp.StatusCode)

	// Get links by group
	resp, arr := doJSONArray(app, "GET", "/api/links-by-group")
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, arr)
}

// ==================== Link Flows CRUD ====================

func TestLinkFlowsCRUD(t *testing.T) {
	app := newTestApp(t)

	// Create group first
	resp, group := doJSON(app, "POST", "/api/link-groups", map[string]interface{}{
		"name": "FlowGroup",
	})
	require.Equal(t, 201, resp.StatusCode)
	gid := group["id"].(string)

	// Create flow
	resp, flow := doJSON(app, "POST", "/api/link-flows", map[string]interface{}{
		"name":     "MyFlow",
		"group_id": gid,
	})
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "MyFlow", flow["name"])
	flowID := flow["id"].(string)

	// List all
	resp, arr := doJSONArray(app, "GET", "/api/link-flows")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, arr, 1)

	// List by group_id
	resp, arr = doJSONArray(app, "GET", "/api/link-flows?group_id="+gid)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, arr, 1)

	// Update flow
	resp, updated := doJSON(app, "PUT", "/api/link-flows/"+flowID, map[string]interface{}{
		"name": "RenamedFlow",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "RenamedFlow", updated["name"])

	// Create link in group
	resp, link := doJSON(app, "POST", "/api/links", map[string]interface{}{
		"name":     "FlowLink",
		"url":      "https://flow.com",
		"group_id": gid,
	})
	require.Equal(t, 201, resp.StatusCode)
	linkID := link["id"].(string)

	// Add link to flow
	resp, _ = doJSON(app, "POST", fmt.Sprintf("/api/link-flows/%s/links", flowID), map[string]interface{}{
		"link_id": linkID,
	})
	assert.Equal(t, 200, resp.StatusCode)

	// Update link order in flow
	resp, _ = doJSON(app, "PUT", fmt.Sprintf("/api/link-flows/%s/links/%s", flowID, linkID), map[string]interface{}{
		"order_index": 5,
	})
	assert.Equal(t, 200, resp.StatusCode)

	// Remove link from flow
	resp, _ = doJSON(app, "DELETE", fmt.Sprintf("/api/link-flows/%s/links/%s", flowID, linkID), nil)
	assert.Equal(t, 200, resp.StatusCode)

	// Delete flow
	resp, _ = doJSON(app, "DELETE", "/api/link-flows/"+flowID, nil)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLinkFlowsValidation(t *testing.T) {
	app := newTestApp(t)

	// Missing fields
	resp, _ := doJSON(app, "POST", "/api/link-flows", map[string]interface{}{
		"name": "NoGroup",
	})
	assert.Equal(t, 400, resp.StatusCode)

	// Invalid group_id filter
	resp, _ = doJSON(app, "GET", "/api/link-flows?group_id=bad-uuid", nil)
	assert.Equal(t, 400, resp.StatusCode)

	// Invalid ID on update
	resp, _ = doJSON(app, "PUT", "/api/link-flows/bad-id", map[string]interface{}{})
	assert.Equal(t, 400, resp.StatusCode)

	// Invalid IDs on link operations
	resp, _ = doJSON(app, "POST", "/api/link-flows/bad-id/links", map[string]interface{}{"link_id": "x"})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "PUT", "/api/link-flows/bad-id/links/bad-id", map[string]interface{}{})
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "DELETE", "/api/link-flows/bad-id/links/bad-id", nil)
	assert.Equal(t, 400, resp.StatusCode)
}

// ==================== Search Engines CRUD ====================

func TestSearchEnginesCRUD(t *testing.T) {
	app := newTestApp(t)

	// List
	resp, arr := doJSONArray(app, "GET", "/api/search-engines")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, arr, 0)

	// Create custom
	resp, created := doJSON(app, "POST", "/api/search-engines", map[string]interface{}{
		"name":         "DuckDuckGo",
		"url_template": "https://duckduckgo.com/?q={query}",
	})
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "DuckDuckGo", created["name"])
	id := fmt.Sprintf("%.0f", created["id"].(float64))

	// Update
	resp, updated := doJSON(app, "PUT", "/api/search-engines/"+id, map[string]interface{}{
		"name": "DDG",
	})
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "DDG", updated["name"])

	// Delete
	resp, _ = doJSON(app, "DELETE", "/api/search-engines/"+id, nil)
	assert.Equal(t, 200, resp.StatusCode)

	resp, arr = doJSONArray(app, "GET", "/api/search-engines")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, arr, 0)
}

func TestSearchEnginesValidation(t *testing.T) {
	app := newTestApp(t)

	// Missing name
	resp, _ := doJSON(app, "POST", "/api/search-engines", map[string]interface{}{
		"url_template": "x",
	})
	assert.Equal(t, 400, resp.StatusCode)

	// Invalid ID on update
	resp, _ = doJSON(app, "PUT", "/api/search-engines/abc", map[string]interface{}{})
	assert.Equal(t, 400, resp.StatusCode)

	// Non-existent ID on update
	resp, _ = doJSON(app, "PUT", "/api/search-engines/99999", map[string]interface{}{"name": "x"})
	assert.Equal(t, 404, resp.StatusCode)

	// Invalid ID on delete
	resp, _ = doJSON(app, "DELETE", "/api/search-engines/abc", nil)
	assert.Equal(t, 400, resp.StatusCode)

	resp, _ = doJSON(app, "POST", "/api/search-engines/seed", nil)
	assert.Equal(t, 405, resp.StatusCode)
}

// ==================== External Requests CRUD & Execute ====================
