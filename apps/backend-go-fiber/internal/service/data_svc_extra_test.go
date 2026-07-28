package service

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func zipWithFile(t *testing.T, name string, data string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(name)
	require.NoError(t, err)
	_, err = f.Write([]byte(data))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

func TestImportRejectsTraversalAndOversizedEntry(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDataSvc(db, t.TempDir())
	_, err := svc.Import(zipWithFile(t, "../settings.json", "[]"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path")

	_, err = svc.Import(zipWithFile(t, "settings.json", strings.Repeat("x", maxBackupEntrySize+1)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

func TestReplaceRejectsLegacyBeforeCreatingBackup(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDataSvc(db, t.TempDir())
	_, err := svc.Import(zipWithFile(t, "settings.json", "[]"), "replace")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "merge only")
}

func TestImportInvalidJSONRollsBack(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDataSvc(db)

	_, err := svc.Import(zipWithFile(t, "settings.json", "{not-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "settings.json")

	var count int64
	require.NoError(t, db.Table("settings").Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
