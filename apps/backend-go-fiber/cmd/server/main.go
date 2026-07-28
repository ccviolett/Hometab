package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"hometab/internal/config"
	"hometab/internal/database"
	"hometab/internal/frontend"
	"hometab/internal/handler"
	"hometab/internal/repository"
	"hometab/internal/router"
	"hometab/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const maxPortSearchAttempts = 100

func main() {
	portFlag := flag.Int("port", 0, "override server port (e.g. -port 52173)")
	noOpenFlag := flag.Bool("no-open", false, "do not open the application in the default browser")
	installFlag := flag.Bool("install", false, "install Hometab as a user login service")
	uninstallFlag := flag.Bool("uninstall", false, "uninstall the Hometab user login service")
	startFlag := flag.Bool("start", false, "start the installed Hometab service")
	stopFlag := flag.Bool("stop", false, "stop the installed Hometab service")
	statusFlag := flag.Bool("status", false, "show installed Hometab service status")
	flag.Parse()

	if handled := handleServiceCommand(serviceCommandFlags{
		Install:   *installFlag,
		Uninstall: *uninstallFlag,
		Start:     *startFlag,
		Stop:      *stopFlag,
		Status:    *statusFlag,
	}); handled {
		return
	}

	cfg := config.Load()

	if *portFlag > 0 {
		cfg.Server.Port = *portFlag
	}

	zerolog.SetGlobalLevel(cfg.LogLevel())
	if cfg.Log.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	db := database.Connect(cfg)
	database.Migrate(db)

	// Repositories
	linkRepo := repository.NewLinkRepo(db)
	linkGroupRepo := repository.NewLinkGroupRepo(db)
	linkFlowRepo := repository.NewLinkFlowRepo(db)
	linkFlowItemRepo := repository.NewLinkFlowItemRepo(db)
	settingRepo := repository.NewSettingRepo(db)
	searchEngineRepo := repository.NewSearchEngineRepo(db)
	externalRequestRepo := repository.NewExternalRequestRepo(db)
	domainIconRepo := repository.NewDomainIconRepo(db)

	// Services
	linkSvc := service.NewLinkSvc(linkRepo, linkFlowRepo, linkFlowItemRepo)
	linkGroupSvc := service.NewLinkGroupSvc(linkGroupRepo, linkRepo, linkFlowRepo, linkFlowItemRepo)
	linkFlowSvc := service.NewLinkFlowSvc(linkFlowRepo, linkFlowItemRepo, linkRepo)
	settingSvc := service.NewSettingSvc(settingRepo)
	searchEngineSvc := service.NewSearchEngineSvc(searchEngineRepo)
	externalExecutionAllowed := service.IsLoopbackHost(cfg.Server.Host) || cfg.ExternalRequests.AllowRemoteExecution
	externalRequestSvc := service.NewExternalRequestSvc(externalRequestRepo, externalExecutionAllowed)
	iconDir := service.DefaultIconDir(cfg.Database.Path)
	domainIconSvc := service.NewDomainIconSvc(domainIconRepo, iconDir)
	dataSvc := service.NewDataSvc(db, iconDir)

	// Handlers
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

	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB for ZIP import
	})

	// CORS: only enabled in dev mode via env var
	if strings.ToLower(os.Getenv("HOME_CORS_ENABLED")) == "true" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,PATCH",
			AllowHeaders: "Content-Type,Authorization",
		}))
		log.Info().Msg("CORS enabled (dev mode)")
	}

	app.Use(fiberlogger.New())
	app.Use(recover.New())

	router.Setup(app, handlers)

	app.Use("/api", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "api endpoint not found",
		})
	})

	// Serve embedded frontend (SPA fallback)
	distFS, _ := fs.Sub(frontend.DistFS, "dist")
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(distFS),
		Browse:       false,
		NotFoundFile: "index.html",
	}))

	listener, actualPort, err := acquireListener(cfg.Server.Host, cfg.Server.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to acquire listen port")
	}
	if actualPort != cfg.Server.Port {
		log.Warn().
			Int("requested_port", cfg.Server.Port).
			Int("actual_port", actualPort).
			Msg("Requested port is unavailable; using next available high port")
	}

	addr := listener.Addr().String()
	url := fmt.Sprintf("http://%s", net.JoinHostPort(browserHost(cfg.Server.Host), strconv.Itoa(actualPort)))
	log.Info().Str("addr", addr).Str("url", url).Msg("Starting Hometab (Go Fiber)")
	if !*noOpenFlag {
		go openBrowserWhenReady(url)
	}
	if err := app.Listener(listener); err != nil {
		log.Fatal().Err(err).Msg("Server stopped")
	}
}

func acquireListener(host string, startPort int) (net.Listener, int, error) {
	if startPort <= 0 || startPort > 65535 {
		return nil, 0, fmt.Errorf("invalid server port: %d", startPort)
	}
	for offset := 0; offset < maxPortSearchAttempts && startPort+offset <= 65535; offset++ {
		port := startPort + offset
		addr := net.JoinHostPort(host, strconv.Itoa(port))
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			return listener, port, nil
		}
		if !isAddressInUse(err) {
			return nil, 0, fmt.Errorf("listen %s: %w", addr, err)
		}
	}
	return nil, 0, fmt.Errorf("no available port found from %d to %d", startPort, min(startPort+maxPortSearchAttempts-1, 65535))
}

func isAddressInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE) || strings.Contains(strings.ToLower(err.Error()), "address already in use")
}

func browserHost(host string) string {
	switch host {
	case "", "0.0.0.0", "::":
		return "127.0.0.1"
	default:
		return host
	}
}
