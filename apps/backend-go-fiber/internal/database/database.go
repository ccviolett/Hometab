package database

import (
	"os"
	"path/filepath"

	"hometab/internal/config"
	"hometab/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) *gorm.DB {
	logLevel := logger.Silent
	if cfg.Log.Level == "debug" {
		logLevel = logger.Info
	}

	if err := ensureDatabaseDir(cfg.Database.Path); err != nil {
		log.Fatal().Err(err).Str("path", cfg.Database.Path).Msg("failed to create database directory")
	}

	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatal().Err(err).Str("path", cfg.Database.Path).Msg("failed to open sqlite database")
	}

	// Enable WAL mode for better concurrent read performance
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	log.Info().Str("path", cfg.Database.Path).Msg("sqlite database connected")
	return db
}

func ensureDatabaseDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.LinkGroup{},
		&model.Link{},
		&model.LinkFlow{},
		&model.LinkFlowItem{},
		&model.Setting{},
		&model.SearchEngine{},
		&model.ExternalRequest{},
		&model.DomainIcon{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to auto-migrate database")
	}
	migrateLegacyLinkFlowItems(db)
	pruneLegacyData(db)
	log.Info().Msg("database migration completed")
}

func migrateLegacyLinkFlowItems(db *gorm.DB) {
	// Keep flow_id for one compatibility window, but LinkFlowItem is the new source of truth.
	db.Exec(`
		INSERT INTO link_flow_items (flow_id, link_id, order_index, created_at, updated_at)
		SELECT links.flow_id, links.id, links.order_index, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM links
		JOIN link_flows ON link_flows.id = links.flow_id
		WHERE links.flow_id IS NOT NULL AND links.flow_id != ''
		  AND NOT EXISTS (
			SELECT 1 FROM link_flow_items
			WHERE link_flow_items.flow_id = links.flow_id
			  AND link_flow_items.link_id = links.id
		  )
	`)
}

func pruneLegacyData(db *gorm.DB) {
	db.Exec("UPDATE links SET group_id = NULL WHERE group_id = ''")
	db.Exec(`
		DELETE FROM link_groups
		WHERE NOT EXISTS (SELECT 1 FROM links WHERE links.group_id = link_groups.id)
		  AND NOT EXISTS (SELECT 1 FROM link_flows WHERE link_flows.group_id = link_groups.id)
	`)
	db.Exec("DROP TABLE IF EXISTS services")
	dropColumnIfExists(db, "links", "favicon")
	pruneSearchEngineActiveColumn(db)
	db.Exec("DELETE FROM search_engines WHERE name = '' OR url_template = ''")
}

func pruneSearchEngineActiveColumn(db *gorm.DB) {
	if !columnExists(db, "search_engines", "is_active") {
		return
	}
	db.Exec("DELETE FROM search_engines WHERE is_active = 0")
	dropColumnIfExists(db, "search_engines", "is_active")
}

func dropColumnIfExists(db *gorm.DB, table, column string) {
	if !columnExists(db, table, column) {
		return
	}
	db.Exec("ALTER TABLE " + table + " DROP COLUMN " + column)
}

func columnExists(db *gorm.DB, table, column string) bool {
	var count int
	db.Raw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", table, column).Scan(&count)
	return count > 0
}
