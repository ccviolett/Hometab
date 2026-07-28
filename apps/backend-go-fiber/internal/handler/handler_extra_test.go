package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"hometab/internal/testutil"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ==================== Import Missing File ====================

func TestImportMissingFile(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("POST", "/api/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestImportInvalidZip(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "bad.zip")
	part.Write([]byte("not a zip file"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

// ==================== Deep Link Flow Handler Tests ====================

func TestLinkFlowAddLinkMissingLinkID(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	// Create group and flow
	_, group := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	gid := group["id"].(string)
	_, flow := doJSON2(app, "POST", "/api/link-flows", map[string]interface{}{"name": "F", "group_id": gid})
	fid := flow["id"].(string)

	// Missing link_id
	resp, _ := doJSON2(app, "POST", fmt.Sprintf("/api/link-flows/%s/links", fid), map[string]interface{}{})
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkFlowDeleteWithBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	_, group := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	gid := group["id"].(string)
	_, flow := doJSON2(app, "POST", "/api/link-flows", map[string]interface{}{"name": "F", "group_id": gid})
	fid := flow["id"].(string)

	// Delete with keep IDs body
	resp, _ := doJSON2(app, "DELETE", "/api/link-flows/"+fid, map[string]interface{}{
		"link_ids_to_keep": []string{},
	})
	assert.Equal(t, 200, resp.StatusCode)
}

// ==================== Handler Invalid Body Tests ====================

func TestLinkUpdateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	_, group := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	gid := group["id"].(string)
	_, link := doJSON2(app, "POST", "/api/links", map[string]interface{}{"name": "L", "url": "https://a.com", "group_id": gid})
	lid := link["id"].(string)

	req := httptest.NewRequest("PUT", "/api/links/"+lid, bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkGroupUpdateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	_, g := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	id := g["id"].(string)

	req := httptest.NewRequest("PUT", "/api/link-groups/"+id, bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkFlowCreateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("POST", "/api/link-flows", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkFlowUpdateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	_, group := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	gid := group["id"].(string)
	_, flow := doJSON2(app, "POST", "/api/link-flows", map[string]interface{}{"name": "F", "group_id": gid})
	fid := flow["id"].(string)

	req := httptest.NewRequest("PUT", "/api/link-flows/"+fid, bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkFlowAddLinkInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	_, group := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	gid := group["id"].(string)
	_, flow := doJSON2(app, "POST", "/api/link-flows", map[string]interface{}{"name": "F", "group_id": gid})
	fid := flow["id"].(string)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/link-flows/%s/links", fid), bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkFlowUpdateLinkInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	_, group := doJSON2(app, "POST", "/api/link-groups", map[string]interface{}{"name": "G"})
	gid := group["id"].(string)
	_, flow := doJSON2(app, "POST", "/api/link-flows", map[string]interface{}{"name": "F", "group_id": gid})
	fid := flow["id"].(string)
	_, link := doJSON2(app, "POST", "/api/links", map[string]interface{}{"name": "L", "url": "https://a.com", "group_id": gid})
	lid := link["id"].(string)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/link-flows/%s/links/%s", fid, lid), bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSearchEngineCreateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("POST", "/api/search-engines", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSearchEngineUpdateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("PUT", "/api/search-engines/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSettingsCreateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("POST", "/api/settings", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSettingsUpdateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("PUT", "/api/settings/theme", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkDeleteInvalidID(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	resp, _ := doJSON2(app, "DELETE", "/api/links/bad-uuid", nil)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkGroupDeleteInvalidID(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	resp, _ := doJSON2(app, "DELETE", "/api/link-groups/bad-uuid", nil)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkFlowDeleteInvalidID(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	resp, _ := doJSON2(app, "DELETE", "/api/link-flows/bad-uuid", nil)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkUpdateInvalidID(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	resp, _ := doJSON2(app, "PUT", "/api/links/bad-uuid", map[string]interface{}{"name": "x"})
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkGroupUpdateInvalidID(t *testing.T) {
	app, _ := testutil.SetupTestApp()
	resp, _ := doJSON2(app, "PUT", "/api/link-groups/bad-uuid", map[string]interface{}{"name": "x"})
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkGroupCreateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("POST", "/api/link-groups", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLinkCreateInvalidBody(t *testing.T) {
	app, _ := testutil.SetupTestApp()

	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	assert.Equal(t, 400, resp.StatusCode)
}

// helper (same as in handler_test.go but avoids name collision)
func doJSON2(app *fiber.App, method, url string, body interface{}) (*http.Response, map[string]interface{}) {
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
