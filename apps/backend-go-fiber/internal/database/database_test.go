package database

import (
	"testing"

	"hometab/internal/config"
	"hometab/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectAndMigrate(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: ":memory:"},
		Log:      config.LogConfig{Level: "info"},
	}
	db := Connect(cfg)
	assert.NotNil(t, db)

	Migrate(db)

	// Verify tables exist
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Scan(&tables)
	assert.Contains(t, tables, "links")
	assert.Contains(t, tables, "link_groups")
	assert.Contains(t, tables, "link_flows")
	assert.Contains(t, tables, "link_flow_items")
	assert.Contains(t, tables, "settings")
	assert.Contains(t, tables, "search_engines")
}

func TestMigrateLegacyLinkFlowItemsIsIdempotent(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: ":memory:"},
		Log:      config.LogConfig{Level: "info"},
	}
	db := Connect(cfg)
	Migrate(db)

	group := model.LinkGroup{Name: "Group"}
	require.NoError(t, db.Create(&group).Error)
	flow := model.LinkFlow{GroupID: group.ID, Name: "Flow"}
	require.NoError(t, db.Create(&flow).Error)
	link := model.Link{Name: "Link", URL: "https://example.com", GroupID: &group.ID, FlowID: &flow.ID, OrderIndex: 30}
	require.NoError(t, db.Create(&link).Error)

	Migrate(db)
	Migrate(db)

	var items []model.LinkFlowItem
	require.NoError(t, db.Find(&items).Error)
	require.Len(t, items, 1)
	assert.Equal(t, flow.ID, items[0].FlowID)
	assert.Equal(t, link.ID, items[0].LinkID)
	assert.Equal(t, 30, items[0].OrderIndex)
}

func TestConnectDebugLevel(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: ":memory:"},
		Log:      config.LogConfig{Level: "debug"},
	}
	db := Connect(cfg)
	assert.NotNil(t, db)
}

func TestMigratePrunesLegacyData(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: ":memory:"},
		Log:      config.LogConfig{Level: "info"},
	}
	db := Connect(cfg)
	require.NotNil(t, db)

	require.NoError(t, db.Exec(`CREATE TABLE services (id TEXT PRIMARY KEY, name TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE links (id TEXT PRIMARY KEY, name TEXT NOT NULL, url TEXT NOT NULL, favicon TEXT)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE search_engines (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, url_template TEXT NOT NULL, is_active BOOLEAN DEFAULT TRUE)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE settings (id INTEGER PRIMARY KEY AUTOINCREMENT, key TEXT NOT NULL UNIQUE, value_json TEXT NOT NULL)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO search_engines (name, url_template, is_active) VALUES ('Old Active', 'https://active.test?q={query}', 1), ('Old Inactive', 'https://inactive.test?q={query}', 0)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO settings (key, value_json) VALUES ('theme', '"dark"')`).Error)

	Migrate(db)

	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Scan(&tables)
	assert.NotContains(t, tables, "services")

	assert.False(t, columnExists(db, "links", "favicon"))
	assert.False(t, columnExists(db, "search_engines", "is_active"))

	var engineNames []string
	db.Raw("SELECT name FROM search_engines ORDER BY name").Scan(&engineNames)
	assert.Equal(t, []string{"Old Active"}, engineNames)

	var keys []string
	db.Raw("SELECT key FROM settings ORDER BY key").Scan(&keys)
	assert.Equal(t, []string{"theme"}, keys)
}
