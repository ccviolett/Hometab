package service

import (
	"os"
	"path/filepath"
	"testing"

	"hometab/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var onePixelPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89,
}

func TestDomainIconUploadListAndDelete(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	svc := NewDomainIconSvc(repository.NewDomainIconRepo(db), dir)

	item, err := svc.Upload("example.com", onePixelPNG)
	require.NoError(t, err)
	assert.Equal(t, "manual", item.Source)
	assert.Equal(t, "ready", item.Status)
	_, err = os.Stat(filepath.Join(dir, item.IconPath))
	require.NoError(t, err)

	items, err := svc.List()
	require.NoError(t, err)
	require.Len(t, items, 1)

	path := items[0].IconPath
	require.NoError(t, svc.Delete("example.com"))
	_, err = os.Stat(filepath.Join(dir, path))
	assert.True(t, os.IsNotExist(err))
	items, err = svc.List()
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestDomainIconUploadValidation(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainIconSvc(repository.NewDomainIconRepo(db), t.TempDir())
	_, err := svc.Upload("../bad", onePixelPNG)
	assert.Error(t, err)
	_, err = svc.Upload("example.com", nil)
	assert.Error(t, err)
	_, err = svc.Upload("example.com", []byte("not an image"))
	assert.Error(t, err)
	_, err = svc.Upload("example.com", make([]byte, 512*1024+1))
	assert.Error(t, err)
	_, err = svc.Upload("example.com", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	assert.Error(t, err)

	item, err := svc.Upload("example.com", []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`))
	require.NoError(t, err)
	assert.Equal(t, "image/svg+xml", item.ContentType)
}
