package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"hometab/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxBackupEntries       = 1000
	maxBackupEntrySize     = 2 << 20
	maxBackupUnpackedBytes = 100 << 20
)

var requiredV2Files = []string{
	"metadata.json", "link_groups.json", "link_flows.json", "link_flow_items.json",
	"links.json", "settings.json", "search_engines.json", "external_requests.json", "domain_icons.json",
}

func (s *DataSvc) Import(zipData []byte, modes ...string) (*ImportResult, error) {
	mode := "merge"
	if len(modes) > 0 && modes[0] != "" {
		mode = modes[0]
	}
	if mode != "merge" && mode != "replace" {
		return nil, fmt.Errorf("invalid import mode")
	}
	snapshot, err := parseBackup(zipData, mode)
	if err != nil {
		return nil, err
	}
	if err := validateSnapshot(snapshot); err != nil {
		return nil, err
	}
	result := newImportResult(mode)
	if mode == "replace" {
		backup, err := s.createPreRestoreBackup()
		if err != nil {
			return nil, fmt.Errorf("create pre-restore backup: %w", err)
		}
		result.PreRestore = backup
	}
	if err := s.applySnapshot(snapshot, mode, result); err != nil {
		result.Errors = append(result.Errors, ImportErrorDetail{File: "backup.zip", Reason: err.Error()})
		return result, err
	}
	return result, nil
}

func parseBackup(data []byte, mode string) (*backupSnapshot, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid zip file: %w", err)
	}
	if len(reader.File) == 0 || len(reader.File) > maxBackupEntries {
		return nil, fmt.Errorf("invalid backup entry count")
	}
	files := make(map[string][]byte)
	icons := make(map[string][]byte)
	var total uint64
	for _, file := range reader.File {
		clean := filepath.ToSlash(filepath.Clean(file.Name))
		if file.FileInfo().IsDir() {
			continue
		}
		if clean == "." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") || strings.Contains(clean, "/../") {
			return nil, fmt.Errorf("invalid backup path: %s", file.Name)
		}
		if file.UncompressedSize64 > maxBackupEntrySize {
			return nil, fmt.Errorf("backup entry too large: %s", file.Name)
		}
		total += file.UncompressedSize64
		if total > maxBackupUnpackedBytes {
			return nil, fmt.Errorf("backup unpacked size exceeds limit")
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", file.Name, err)
		}
		content, readErr := io.ReadAll(io.LimitReader(rc, maxBackupEntrySize+1))
		closeErr := rc.Close()
		if readErr != nil || closeErr != nil || len(content) > maxBackupEntrySize {
			return nil, fmt.Errorf("read backup entry: %s", file.Name)
		}
		if strings.HasPrefix(clean, "icons/") && strings.Count(clean, "/") == 1 {
			base := filepath.Base(clean)
			if _, exists := icons[base]; exists {
				return nil, fmt.Errorf("duplicate icon: %s", base)
			}
			icons[base] = content
			continue
		}
		if strings.Contains(clean, "/") || filepath.Base(clean) != clean {
			return nil, fmt.Errorf("unsupported backup path: %s", file.Name)
		}
		if _, exists := files[clean]; exists {
			return nil, fmt.Errorf("duplicate backup file: %s", clean)
		}
		files[clean] = content
	}

	snapshot := &backupSnapshot{IconFiles: icons}
	if metadata, ok := files["metadata.json"]; ok {
		if err := json.Unmarshal(metadata, &snapshot.Metadata); err != nil {
			return nil, fmt.Errorf("metadata.json: %w", err)
		}
	}
	version := snapshot.Metadata.FormatVersion
	if version == "" {
		version = snapshot.Metadata.Version
	}
	if version == "" {
		version = "1.2.0"
	}
	if strings.HasPrefix(version, "2.") {
		for _, name := range requiredV2Files {
			if _, ok := files[name]; !ok {
				return nil, fmt.Errorf("missing required file: %s", name)
			}
		}
	} else if mode == "replace" {
		return nil, fmt.Errorf("legacy backup %s supports merge only", version)
	} else if !strings.HasPrefix(version, "1.") {
		return nil, fmt.Errorf("unsupported backup format version: %s", version)
	}
	snapshot.Metadata.FormatVersion = version

	decode := func(name string, target interface{}) error {
		content, ok := files[name]
		if !ok {
			return nil
		}
		if err := json.Unmarshal(content, target); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}
	if err := decode("link_groups.json", &snapshot.Groups); err != nil {
		return nil, err
	}
	if err := decode("link_flows.json", &snapshot.Flows); err != nil {
		return nil, err
	}
	if err := decode("link_flow_items.json", &snapshot.FlowItems); err != nil {
		return nil, err
	}
	if err := decode("links.json", &snapshot.Links); err != nil {
		return nil, err
	}
	if err := decode("settings.json", &snapshot.Settings); err != nil {
		var legacy []model.Setting
		if legacyErr := decode("settings.json", &legacy); legacyErr != nil {
			return nil, err
		}
		for _, item := range legacy {
			snapshot.Settings = append(snapshot.Settings, exportSetting{Key: item.Key, Value: item.GetValue()})
		}
	}
	if err := decode("search_engines.json", &snapshot.Engines); err != nil {
		return nil, err
	}
	if err := decode("external_requests.json", &snapshot.ExternalRequests); err != nil {
		return nil, err
	}
	if err := decode("domain_icons.json", &snapshot.DomainIcons); err != nil {
		return nil, err
	}

	if len(snapshot.FlowItems) == 0 {
		for _, link := range snapshot.Links {
			if link.FlowID != nil {
				snapshot.FlowItems = append(snapshot.FlowItems, model.LinkFlowItem{FlowID: *link.FlowID, LinkID: link.ID, OrderIndex: link.OrderIndex})
			}
		}
	}
	return snapshot, nil
}

func validateSnapshot(snapshot *backupSnapshot) error {
	groupIDs := make(map[string]bool)
	flowIDs := make(map[string]bool)
	linkIDs := make(map[string]bool)
	for _, group := range snapshot.Groups {
		if group.ID.String() == "00000000-0000-0000-0000-000000000000" {
			return errors.New("link_groups.json: empty id")
		}
		groupIDs[group.ID.String()] = true
	}
	for _, flow := range snapshot.Flows {
		if !groupIDs[flow.GroupID.String()] {
			return fmt.Errorf("link_flows.json: missing group %s", flow.GroupID)
		}
		flowIDs[flow.ID.String()] = true
	}
	for _, link := range snapshot.Links {
		if link.GroupID != nil && !groupIDs[link.GroupID.String()] {
			return fmt.Errorf("links.json: missing group %s", link.GroupID)
		}
		linkIDs[link.ID.String()] = true
	}
	seenItems := make(map[string]bool)
	for _, item := range snapshot.FlowItems {
		key := item.FlowID.String() + "/" + item.LinkID.String()
		if seenItems[key] {
			return fmt.Errorf("link_flow_items.json: duplicate relation %s", key)
		}
		seenItems[key] = true
		if !flowIDs[item.FlowID.String()] || !linkIDs[item.LinkID.String()] {
			return fmt.Errorf("link_flow_items.json: invalid relation %s", key)
		}
	}
	for _, icon := range snapshot.DomainIcons {
		if strings.TrimSpace(icon.Host) == "" {
			return errors.New("domain_icons.json: empty host")
		}
		for _, rel := range []string{icon.IconPath, icon.PendingPath} {
			if rel != "" {
				if filepath.Base(rel) != rel {
					return fmt.Errorf("domain_icons.json: invalid icon path %s", rel)
				}
				if _, ok := snapshot.IconFiles[rel]; !ok {
					return fmt.Errorf("domain_icons.json: missing icon %s", rel)
				}
			}
		}
	}
	return nil
}

func (s *DataSvc) applySnapshot(snapshot *backupSnapshot, mode string, result *ImportResult) error {
	stageDir, err := os.MkdirTemp(filepath.Dir(s.iconDir), ".hometab-icons-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stageDir)
	for name, data := range snapshot.IconFiles {
		if err := os.WriteFile(filepath.Join(stageDir, filepath.Base(name)), data, 0o600); err != nil {
			return err
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if mode == "replace" {
			for _, entity := range []interface{}{&model.LinkFlowItem{}, &model.Link{}, &model.LinkFlow{}, &model.LinkGroup{}, &model.Setting{}, &model.SearchEngine{}, &model.ExternalRequest{}, &model.DomainIcon{}} {
				if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(entity).Error; err != nil {
					return err
				}
			}
		}
		create := func(table string, items interface{}, count int) error {
			if count == 0 {
				return nil
			}
			if mode == "merge" {
				resultDB := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(items)
				if resultDB.Error != nil {
					return resultDB.Error
				}
				result.Imported[table] += int(resultDB.RowsAffected)
				result.Skipped[table] += count - int(resultDB.RowsAffected)
				return nil
			}
			if err := tx.Create(items).Error; err != nil {
				return err
			}
			result.Imported[table] += count
			return nil
		}
		if err := create("link_groups", &snapshot.Groups, len(snapshot.Groups)); err != nil {
			return err
		}
		if err := create("link_flows", &snapshot.Flows, len(snapshot.Flows)); err != nil {
			return err
		}
		if err := create("links", &snapshot.Links, len(snapshot.Links)); err != nil {
			return err
		}
		if err := create("link_flow_items", &snapshot.FlowItems, len(snapshot.FlowItems)); err != nil {
			return err
		}
		settings := make([]model.Setting, 0, len(snapshot.Settings))
		for _, item := range snapshot.Settings {
			if item.Key != "" && item.Value != nil && !isLegacySettingKey(item.Key) {
				settings = append(settings, model.Setting{Key: item.Key, ValueJSON: model.ValueToJSON(item.Value)})
			}
		}
		if err := create("settings", &settings, len(settings)); err != nil {
			return err
		}
		if err := create("search_engines", &snapshot.Engines, len(snapshot.Engines)); err != nil {
			return err
		}
		if err := create("external_requests", &snapshot.ExternalRequests, len(snapshot.ExternalRequests)); err != nil {
			return err
		}
		if err := create("domain_icons", &snapshot.DomainIcons, len(snapshot.DomainIcons)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	if err := os.MkdirAll(s.iconDir, 0o755); err != nil {
		return err
	}
	if mode == "replace" {
		entries, _ := os.ReadDir(s.iconDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				_ = os.Remove(filepath.Join(s.iconDir, entry.Name()))
			}
		}
	}
	for name := range snapshot.IconFiles {
		data, err := os.ReadFile(filepath.Join(stageDir, name))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(s.iconDir, name), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
