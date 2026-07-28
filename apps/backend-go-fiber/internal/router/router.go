package router

import (
	"hometab/internal/handler"
	"hometab/pkg/buildinfo"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Link            *handler.LinkHandler
	LinkGroup       *handler.LinkGroupHandler
	LinkFlow        *handler.LinkFlowHandler
	Setting         *handler.SettingHandler
	SearchEngine    *handler.SearchEngineHandler
	ExternalRequest *handler.ExternalRequestHandler
	DomainIcon      *handler.DomainIconHandler
	Data            *handler.DataHandler
}

func Setup(app *fiber.App, h *Handlers) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "hometab-go"})
	})
	app.Get("/api/build-info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":    buildinfo.Version,
			"build_time": buildinfo.BuildTime,
			"go_version": buildinfo.GoVersion(),
		})
	})

	api := app.Group("/api")

	// Links
	api.Get("/links", h.Link.List)
	api.Post("/links", h.Link.Create)
	api.Put("/link-groups/:groupID/links/order", h.Link.ReorderGroup)
	api.Get("/links/:id", h.Link.Get)
	api.Put("/links/:id", h.Link.Update)
	api.Delete("/links/:id", h.Link.Delete)

	// Link Groups
	api.Get("/link-groups", h.LinkGroup.List)
	api.Post("/link-groups", h.LinkGroup.Create)
	api.Put("/link-groups/order", h.LinkGroup.Reorder)
	api.Get("/link-groups/:id", h.LinkGroup.Get)
	api.Put("/link-groups/:id", h.LinkGroup.Update)
	api.Delete("/link-groups/:id", h.LinkGroup.Delete)
	api.Get("/links-by-group", h.LinkGroup.ListByGroup)

	// Link Flows
	api.Get("/link-flows", h.LinkFlow.List)
	api.Post("/link-flows", h.LinkFlow.Create)
	api.Put("/link-flows/:id", h.LinkFlow.Update)
	api.Delete("/link-flows/:id", h.LinkFlow.Delete)
	api.Post("/link-flows/:id/links", h.LinkFlow.AddLink)
	api.Put("/link-flows/:id/links/order", h.LinkFlow.ReorderLinks)
	api.Put("/link-flows/:fid/links/:lid", h.LinkFlow.UpdateLink)
	api.Delete("/link-flows/:fid/links/:lid", h.LinkFlow.RemoveLink)

	// Search Engines
	api.Get("/search-engines", h.SearchEngine.List)
	api.Post("/search-engines", h.SearchEngine.Create)
	api.Put("/search-engines/:id", h.SearchEngine.Update)
	api.Delete("/search-engines/:id", h.SearchEngine.Delete)

	// External Requests
	api.Get("/external-requests", h.ExternalRequest.List)
	api.Post("/external-requests", h.ExternalRequest.Create)
	api.Get("/external-requests/:id", h.ExternalRequest.Get)
	api.Put("/external-requests/:id", h.ExternalRequest.Update)
	api.Delete("/external-requests/:id", h.ExternalRequest.Delete)
	api.Post("/external-requests/:id/execute", h.ExternalRequest.Execute)

	// Link Icons
	api.Get("/link-icons", h.DomainIcon.List)
	api.Get("/link-icons/resolve", h.DomainIcon.Resolve)
	api.Get("/link-icons/pending/:host", h.DomainIcon.Pending)
	api.Post("/link-icons/check", h.DomainIcon.Check)
	api.Post("/link-icons/refresh-all", h.DomainIcon.RefreshAll)
	api.Post("/link-icons/:host/upload", h.DomainIcon.Upload)
	api.Post("/link-icons/:host/retry", h.DomainIcon.Retry)
	api.Post("/link-icons/:host/choose", h.DomainIcon.Choose)
	api.Delete("/link-icons/:host", h.DomainIcon.Delete)

	// Settings
	api.Get("/settings", h.Setting.List)
	api.Get("/settings/:key", h.Setting.Get)
	api.Post("/settings", h.Setting.CreateOrUpdate)
	api.Put("/settings/:key", h.Setting.Update)
	api.Delete("/settings/:key", h.Setting.Delete)

	// Data Management
	api.Get("/export", h.Data.Export)
	api.Post("/import", h.Data.Import)
	api.Get("/backups/:id/download", h.Data.DownloadBackup)
}
