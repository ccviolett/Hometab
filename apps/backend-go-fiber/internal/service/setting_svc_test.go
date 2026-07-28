package service

import (
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingSvcCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewSettingRepo(db)
	svc := NewSettingSvc(repo)

	// FindAll empty
	m, err := svc.FindAll()
	require.NoError(t, err)
	assert.Len(t, m, 0)

	// CreateOrUpdate (create)
	created, err := svc.CreateOrUpdate(model.SettingCreate{Key: "theme", Value: "dark"})
	require.NoError(t, err)
	assert.Equal(t, "theme", created.Key)
	assert.Equal(t, "dark", created.Value)

	// FindByKey
	got, err := svc.FindByKey("theme")
	require.NoError(t, err)
	assert.Equal(t, "dark", got.Value)

	// FindByKey not found
	_, err = svc.FindByKey("nonexistent")
	assert.Error(t, err)

	// CreateOrUpdate (update existing)
	updated, err := svc.CreateOrUpdate(model.SettingCreate{Key: "theme", Value: "light"})
	require.NoError(t, err)
	assert.Equal(t, "light", updated.Value)

	// Update
	updated2, err := svc.Update("theme", model.SettingUpdate{Value: "auto"})
	require.NoError(t, err)
	assert.Equal(t, "auto", updated2.Value)

	// Update not found
	_, err = svc.Update("nonexistent", model.SettingUpdate{Value: "x"})
	assert.Error(t, err)

	// FindAll non-empty
	m, err = svc.FindAll()
	require.NoError(t, err)
	assert.Len(t, m, 1)

	// Delete
	err = svc.Delete("theme")
	assert.NoError(t, err)
}
