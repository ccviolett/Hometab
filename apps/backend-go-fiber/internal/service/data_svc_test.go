package service

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"hometab/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	db.AutoMigrate(
		&model.LinkGroup{},
		&model.Link{},
		&model.LinkFlow{},
		&model.LinkFlowItem{},
		&model.Setting{},
		&model.SearchEngine{},
		&model.ExternalRequest{},
		&model.DomainIcon{},
	)
	return db
}

func TestExportEmpty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDataSvc(db)

	buf, filename, err := svc.Export()
	require.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, filename, "hometab_backup_")
	assert.Contains(t, filename, ".zip")
	assert.Greater(t, buf.Len(), 0)
}

func TestExportReplaceRoundTripIncludesFlowMemberships(t *testing.T) {
	db := setupTestDB(t)
	iconDir := t.TempDir()
	svc := NewDataSvc(db, iconDir)

	group := model.LinkGroup{Name: "Group", OrderIndex: 10}
	require.NoError(t, db.Create(&group).Error)
	flowA := model.LinkFlow{GroupID: group.ID, Name: "Morning"}
	flowB := model.LinkFlow{GroupID: group.ID, Name: "Release"}
	require.NoError(t, db.Create(&flowA).Error)
	require.NoError(t, db.Create(&flowB).Error)
	link := model.Link{Name: "GitHub", URL: "https://github.com", GroupID: &group.ID, OrderIndex: 20}
	require.NoError(t, db.Create(&link).Error)
	require.NoError(t, db.Create(&model.LinkFlowItem{FlowID: flowA.ID, LinkID: link.ID, OrderIndex: 10}).Error)
	require.NoError(t, db.Create(&model.LinkFlowItem{FlowID: flowB.ID, LinkID: link.ID, OrderIndex: 30}).Error)

	buf, _, err := svc.Export()
	require.NoError(t, err)
	require.NoError(t, db.Create(&model.LinkGroup{Name: "To be replaced"}).Error)

	result, err := svc.Import(buf.Bytes(), "replace")
	require.NoError(t, err)
	require.NotNil(t, result.PreRestore)

	var groups []model.LinkGroup
	var items []model.LinkFlowItem
	require.NoError(t, db.Order("name").Find(&groups).Error)
	require.NoError(t, db.Order("order_index").Find(&items).Error)
	require.Len(t, groups, 1)
	require.Len(t, items, 2)
	assert.Equal(t, "Group", groups[0].Name)
	assert.Equal(t, []int{10, 30}, []int{items[0].OrderIndex, items[1].OrderIndex})
	_, err = svc.BackupPath(result.PreRestore.ID)
	require.NoError(t, err)
}

func TestPreRestoreBackupRetention(t *testing.T) {
	db := setupTestDB(t)
	iconDir := t.TempDir()
	svc := NewDataSvc(db, iconDir)
	for i := 0; i < 7; i++ {
		_, err := svc.createPreRestoreBackup()
		require.NoError(t, err)
	}
	entries, err := os.ReadDir(svc.backupDir)
	require.NoError(t, err)
	assert.Len(t, entries, 5)
}

func TestExportImportRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDataSvc(db)

	// Seed data
	db.Create(&model.Setting{Key: "theme", ValueJSON: `"dark"`})
	db.Create(&model.LinkGroup{Name: "Group1"})
	db.Create(&model.SearchEngine{Name: "Google", URLTemplate: "https://google.com/search?q={query}"})

	buf, _, err := svc.Export()
	require.NoError(t, err)

	// Import into fresh DB
	db2 := setupTestDB(t)
	svc2 := NewDataSvc(db2)
	result, err := svc2.Import(buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported["settings"])
	assert.Equal(t, 1, result.Imported["link_groups"])
	assert.Equal(t, 1, result.Imported["search_engines"])

	// Import again (duplicates should be skipped)
	buf2, _, _ := svc.Export()
	result2, err := svc2.Import(buf2.Bytes())
	require.NoError(t, err)
	assert.Equal(t, 0, result2.Imported["settings"])
	assert.Equal(t, 1, result2.Skipped["settings"])
}

func TestImportInvalidZip(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDataSvc(db)

	_, err := svc.Import([]byte("not a zip"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid zip")
}

func TestExportImportDomainIcons(t *testing.T) {
	db := setupTestDB(t)
	iconDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(iconDir, "example_com-abc123.svg"), []byte("<svg></svg>"), 0644))
	require.NoError(t, db.Create(&model.DomainIcon{
		Host:        "example.com",
		IconPath:    "example_com-abc123.svg",
		ContentType: "image/svg+xml",
		Hash:        "abc123",
		Source:      "auto",
		Status:      "ready",
	}).Error)

	svc := NewDataSvc(db, iconDir)
	buf, _, err := svc.Export()
	require.NoError(t, err)

	zr, err := zip.NewReader(bytesReaderAt(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	foundMeta := false
	foundFile := false
	for _, f := range zr.File {
		if f.Name == "domain_icons.json" {
			foundMeta = true
		}
		if f.Name == "icons/example_com-abc123.svg" {
			foundFile = true
		}
	}
	assert.True(t, foundMeta)
	assert.True(t, foundFile)

	db2 := setupTestDB(t)
	iconDir2 := t.TempDir()
	svc2 := NewDataSvc(db2, iconDir2)
	result, err := svc2.Import(buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported["domain_icons"])
	restored, err := os.ReadFile(filepath.Join(iconDir2, "example_com-abc123.svg"))
	require.NoError(t, err)
	assert.Equal(t, "<svg></svg>", string(restored))
}

type bytesReaderAt []byte

func (b bytesReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(b)) {
		return 0, io.EOF
	}
	n := copy(p, b[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
