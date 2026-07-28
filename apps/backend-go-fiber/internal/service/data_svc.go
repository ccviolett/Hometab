package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"hometab/internal/model"
	"hometab/pkg/buildinfo"

	"gorm.io/gorm"
)

const backupFormatVersion = "2.0.0"

type DataSvc struct {
	db        *gorm.DB
	iconDir   string
	backupDir string
}

func NewDataSvc(db *gorm.DB, iconDir ...string) *DataSvc {
	dir := ""
	if len(iconDir) > 0 {
		dir = iconDir[0]
	}
	if dir == "" {
		dir = "./icons"
	}
	return &DataSvc{db: db, iconDir: dir, backupDir: filepath.Join(filepath.Dir(dir), "backups")}
}

type exportMetadata struct {
	FormatVersion string         `json:"format_version"`
	AppVersion    string         `json:"app_version"`
	ExportedAt    string         `json:"exported_at"`
	Tables        map[string]int `json:"tables"`
	// Version keeps legacy readers and 1.2.0 fixtures understandable.
	Version string `json:"version,omitempty"`
}

type exportSetting struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ImportErrorDetail struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

type BackupReference struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

type ImportResult struct {
	Imported      map[string]int      `json:"imported"`
	Skipped       map[string]int      `json:"skipped"`
	Errors        []ImportErrorDetail `json:"errors"`
	PreRestore    *BackupReference    `json:"pre_restore_backup,omitempty"`
	FormatVersion string              `json:"format_version"`
	RequestedMode string              `json:"mode"`
}

func newImportResult(mode string) *ImportResult {
	return &ImportResult{
		Imported:      make(map[string]int),
		Skipped:       make(map[string]int),
		Errors:        make([]ImportErrorDetail, 0),
		FormatVersion: backupFormatVersion,
		RequestedMode: mode,
	}
}

type backupSnapshot struct {
	Metadata         exportMetadata
	Links            []model.Link
	Groups           []model.LinkGroup
	Flows            []model.LinkFlow
	FlowItems        []model.LinkFlowItem
	Settings         []exportSetting
	Engines          []model.SearchEngine
	ExternalRequests []model.ExternalRequest
	DomainIcons      []model.DomainIcon
	IconFiles        map[string][]byte
}

func (s *DataSvc) Export() (*bytes.Buffer, string, error) {
	snapshot, err := s.snapshot()
	if err != nil {
		return nil, "", err
	}
	buf, err := writeBackup(snapshot)
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("hometab_backup_%s.zip", time.Now().Format("20060102_150405"))
	return buf, filename, nil
}

func (s *DataSvc) snapshot() (*backupSnapshot, error) {
	snapshot := &backupSnapshot{IconFiles: make(map[string][]byte)}
	if err := s.db.Find(&snapshot.Links).Error; err != nil {
		return nil, err
	}
	if err := s.db.Find(&snapshot.Groups).Error; err != nil {
		return nil, err
	}
	if err := s.db.Find(&snapshot.Flows).Error; err != nil {
		return nil, err
	}
	if err := s.db.Find(&snapshot.FlowItems).Error; err != nil {
		return nil, err
	}
	var settings []model.Setting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	for _, item := range settings {
		if !isLegacySettingKey(item.Key) {
			snapshot.Settings = append(snapshot.Settings, exportSetting{Key: item.Key, Value: item.GetValue()})
		}
	}
	if err := s.db.Find(&snapshot.Engines).Error; err != nil {
		return nil, err
	}
	if err := s.db.Find(&snapshot.ExternalRequests).Error; err != nil {
		return nil, err
	}
	if err := s.db.Find(&snapshot.DomainIcons).Error; err != nil {
		return nil, err
	}
	for _, icon := range snapshot.DomainIcons {
		for _, rel := range []string{icon.IconPath, icon.PendingPath} {
			if rel == "" {
				continue
			}
			base := filepath.Base(rel)
			if _, exists := snapshot.IconFiles[base]; exists {
				continue
			}
			data, err := os.ReadFile(filepath.Join(s.iconDir, base))
			if err != nil {
				return nil, fmt.Errorf("read icon %s: %w", base, err)
			}
			snapshot.IconFiles[base] = data
		}
	}
	snapshot.Metadata = exportMetadata{
		FormatVersion: backupFormatVersion,
		AppVersion:    buildinfo.Version,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Version:       backupFormatVersion,
		Tables: map[string]int{
			"links":             len(snapshot.Links),
			"link_groups":       len(snapshot.Groups),
			"link_flows":        len(snapshot.Flows),
			"link_flow_items":   len(snapshot.FlowItems),
			"settings":          len(snapshot.Settings),
			"search_engines":    len(snapshot.Engines),
			"external_requests": len(snapshot.ExternalRequests),
			"domain_icons":      len(snapshot.DomainIcons),
		},
	}
	return snapshot, nil
}

func writeBackup(snapshot *backupSnapshot) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	writeJSON := func(name string, value interface{}) error {
		entry, err := w.Create(name)
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		_, err = entry.Write(data)
		return err
	}
	files := []struct {
		name  string
		value interface{}
	}{
		{"metadata.json", snapshot.Metadata},
		{"link_groups.json", snapshot.Groups},
		{"link_flows.json", snapshot.Flows},
		{"link_flow_items.json", snapshot.FlowItems},
		{"links.json", snapshot.Links},
		{"settings.json", snapshot.Settings},
		{"search_engines.json", snapshot.Engines},
		{"external_requests.json", snapshot.ExternalRequests},
		{"domain_icons.json", snapshot.DomainIcons},
	}
	for _, file := range files {
		if err := writeJSON(file.name, file.value); err != nil {
			_ = w.Close()
			return nil, err
		}
	}
	iconNames := make([]string, 0, len(snapshot.IconFiles))
	for name := range snapshot.IconFiles {
		iconNames = append(iconNames, name)
	}
	sort.Strings(iconNames)
	for _, name := range iconNames {
		entry, err := w.Create(filepath.ToSlash(filepath.Join("icons", filepath.Base(name))))
		if err != nil {
			_ = w.Close()
			return nil, err
		}
		if _, err := entry.Write(snapshot.IconFiles[name]); err != nil {
			_ = w.Close()
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *DataSvc) createPreRestoreBackup() (*BackupReference, error) {
	buf, filename, err := s.Export()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(s.backupDir, 0o700); err != nil {
		return nil, err
	}
	id := fmt.Sprintf("pre_restore_%d", time.Now().UnixNano())
	storedName := id + ".zip"
	path := filepath.Join(s.backupDir, storedName)
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return nil, err
	}
	if err := s.pruneBackups(5); err != nil {
		return nil, err
	}
	return &BackupReference{ID: id, Filename: filename}, nil
}

func (s *DataSvc) pruneBackups(keep int) error {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return err
	}
	files := make([]os.DirEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".zip" {
			files = append(files, entry)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name() > files[j].Name() })
	if len(files) <= keep {
		return nil
	}
	for _, entry := range files[keep:] {
		if err := os.Remove(filepath.Join(s.backupDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataSvc) BackupPath(id string) (string, error) {
	if id == "" || filepath.Base(id) != id {
		return "", fmt.Errorf("invalid backup id")
	}
	path := filepath.Join(s.backupDir, id+".zip")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func isLegacySettingKey(string) bool { return false }
