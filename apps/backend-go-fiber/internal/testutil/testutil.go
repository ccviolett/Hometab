package testutil

import (
	"os"
	"path/filepath"

	"hometab/internal/handler"
	"hometab/internal/model"
	"hometab/internal/repository"
	"hometab/internal/router"
	"hometab/internal/service"
	"hometab/pkg/buildinfo"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to open test database: " + err.Error())
	}
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

func SetupTestApp() (*fiber.App, *gorm.DB) {
	db := SetupTestDB()
	iconDir, err := os.MkdirTemp("", "hometab-test-icons-")
	if err != nil {
		panic("failed to create test icon directory: " + err.Error())
	}
	linkRepo := repository.NewLinkRepo(db)
	linkGroupRepo := repository.NewLinkGroupRepo(db)
	linkFlowRepo := repository.NewLinkFlowRepo(db)
	linkFlowItemRepo := repository.NewLinkFlowItemRepo(db)
	settingRepo := repository.NewSettingRepo(db)
	searchEngineRepo := repository.NewSearchEngineRepo(db)
	externalRequestRepo := repository.NewExternalRequestRepo(db)
	domainIconRepo := repository.NewDomainIconRepo(db)

	linkSvc := service.NewLinkSvc(linkRepo, linkFlowRepo, linkFlowItemRepo)
	linkGroupSvc := service.NewLinkGroupSvc(linkGroupRepo, linkRepo, linkFlowRepo, linkFlowItemRepo)
	linkFlowSvc := service.NewLinkFlowSvc(linkFlowRepo, linkFlowItemRepo, linkRepo)
	settingSvc := service.NewSettingSvc(settingRepo)
	searchEngineSvc := service.NewSearchEngineSvc(searchEngineRepo)
	externalRequestSvc := service.NewExternalRequestSvc(externalRequestRepo)
	domainIconSvc := service.NewDomainIconSvc(domainIconRepo, iconDir)
	dataSvc := service.NewDataSvc(db, iconDir)

	handlers := &router.Handlers{
		Link:            handler.NewLinkHandler(linkSvc),
		LinkGroup:       handler.NewLinkGroupHandler(linkGroupSvc),
		LinkFlow:        handler.NewLinkFlowHandler(linkFlowSvc),
		Setting:         handler.NewSettingHandler(settingSvc),
		SearchEngine:    handler.NewSearchEngineHandler(searchEngineSvc),
		ExternalRequest: handler.NewExternalRequestHandler(externalRequestSvc),
		DomainIcon:      handler.NewDomainIconHandler(domainIconSvc, linkSvc),
		Data:            handler.NewDataHandler(dataSvc),
	}

	buildinfo.Version = "test"
	buildinfo.BuildTime = "2025-01-01 00:00:00 UTC"

	app := fiber.New(fiber.Config{BodyLimit: 50 * 1024 * 1024})
	router.Setup(app, handlers)
	return app, db
}

func tTempIconDir() string {
	return filepath.Join(os.TempDir(), "home-test-icons")
}
