package service

import (
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchEngineSvcCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewSearchEngineRepo(db)
	svc := NewSearchEngineSvc(repo)

	// FindAll empty
	items, err := svc.FindAll()
	require.NoError(t, err)
	assert.Len(t, items, 0)

	created, err := svc.Create(model.SearchEngineCreate{
		Name:        "DDG",
		URLTemplate: "https://duckduckgo.com/?q={query}",
	})
	require.NoError(t, err)

	created2, err := svc.Create(model.SearchEngineCreate{
		Name:        "Inactive",
		URLTemplate: "https://x.com/?q={query}",
		Icon:        strPtr("search"),
		Description: strPtr("Test"),
		Color:       strPtr("#000"),
	})
	require.NoError(t, err)
	_ = created2 // isActive behavior depends on GORM default handling

	// FindByID
	got, err := svc.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "DDG", got.Name)

	// Update all fields
	newName := "DuckDuckGo"
	newURL := "https://ddg.gg/?q={query}"
	newIcon := "duck"
	newDesc := "Privacy search"
	newColor := "#FF0000"
	updated, err := svc.Update(created.ID, model.SearchEngineUpdate{
		Name:        &newName,
		URLTemplate: &newURL,
		Icon:        &newIcon,
		Description: &newDesc,
		Color:       &newColor,
	})
	require.NoError(t, err)
	assert.Equal(t, "DuckDuckGo", updated.Name)
	assert.Equal(t, "https://ddg.gg/?q={query}", updated.URLTemplate)

	// Update not found
	_, err = svc.Update(99999, model.SearchEngineUpdate{Name: &newName})
	assert.Error(t, err)

	// Delete
	err = svc.Delete(created.ID)
	assert.NoError(t, err)
	_, err = svc.FindByID(created.ID)
	assert.Error(t, err)
}

func TestSearchEngineSvcFindByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewSearchEngineRepo(db)
	svc := NewSearchEngineSvc(repo)

	_, err := svc.FindByID(99999)
	assert.Error(t, err)
}

func strPtr(s string) *string { return &s }
