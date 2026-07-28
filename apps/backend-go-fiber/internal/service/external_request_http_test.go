package service

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalAddressPolicy(t *testing.T) {
	allowed := []string{"127.0.0.1", "10.0.0.1", "192.168.1.10", "8.8.8.8", "::1", "fd00::1"}
	for _, raw := range allowed {
		assert.False(t, externalAddressDenied(netip.MustParseAddr(raw)), raw)
	}
	denied := []string{"0.0.0.0", "169.254.169.254", "224.0.0.1", "255.255.255.255", "::", "fe80::1", "ff02::1"}
	for _, raw := range denied {
		assert.True(t, externalAddressDenied(netip.MustParseAddr(raw)), raw)
	}
}

func TestIsLoopbackHost(t *testing.T) {
	assert.True(t, IsLoopbackHost("127.0.0.1"))
	assert.True(t, IsLoopbackHost("::1"))
	assert.True(t, IsLoopbackHost("localhost"))
	assert.False(t, IsLoopbackHost("0.0.0.0"))
	assert.False(t, IsLoopbackHost("192.168.1.10"))
}

func TestExternalRequestExecutionGate(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewExternalRequestRepo(db)
	item := model.ExternalRequest{Name: "Local", Method: "GET", URL: "http://127.0.0.1:1", Enabled: true}
	require.NoError(t, repo.Create(&item))

	svc := NewExternalRequestSvc(repo, false)
	_, err := svc.Execute(item.ID)
	require.Error(t, err)
	assert.Equal(t, "execution_disabled", err.Error())
}

func TestExternalRequestLoopbackAndResponseLimit(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", int(externalRequestMaxBodyBytes)+1)))
	}))
	defer target.Close()
	db := setupTestDB(t)
	repo := repository.NewExternalRequestRepo(db)
	item := model.ExternalRequest{Name: "Large", Method: "GET", URL: target.URL, Enabled: true}
	require.NoError(t, repo.Create(&item))

	result, err := NewExternalRequestSvc(repo, true).Execute(item.ID)
	require.NoError(t, err)
	assert.Equal(t, "response_too_large", result.Error)
}

func TestExternalRequestRedirectValidation(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data", http.StatusFound)
	}))
	defer target.Close()
	db := setupTestDB(t)
	repo := repository.NewExternalRequestRepo(db)
	item := model.ExternalRequest{Name: "Redirect", Method: "GET", URL: target.URL, Enabled: true}
	require.NoError(t, repo.Create(&item))

	result, err := NewExternalRequestSvc(repo, true).Execute(item.ID)
	require.NoError(t, err)
	assert.Equal(t, "target_denied", result.Error)
}

func TestExternalRequestConcurrencyLimit(t *testing.T) {
	h := newExternalRequestHTTP()
	for i := 0; i < cap(h.sem); i++ {
		h.sem <- struct{}{}
	}
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	_, err := h.do(req)
	require.Error(t, err)
	assert.Equal(t, "concurrency_limited", err.Error())
}

func TestExternalRequestRejectsUserInfoAndInvalidJSONBody(t *testing.T) {
	db := setupTestDB(t)
	svc := NewExternalRequestSvc(repository.NewExternalRequestRepo(db))
	_, err := svc.Create(model.ExternalRequestCreate{Name: "Secret", URL: "https://user:pass@example.com"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "pass")

	_, err = svc.Create(model.ExternalRequestCreate{Name: "JSON", URL: "https://example.com", Method: "POST", BodyType: "json", Body: "{"})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid json body"))
}
