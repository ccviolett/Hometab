package handler

import (
	"io"
	"os"
	"path/filepath"

	"hometab/internal/service"
	"hometab/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type DataHandler struct {
	svc *service.DataSvc
}

func NewDataHandler(svc *service.DataSvc) *DataHandler {
	return &DataHandler{svc: svc}
}

func (h *DataHandler) Export(c *fiber.Ctx) error {
	buf, filename, err := h.svc.Export()
	if err != nil {
		return response.InternalError(c, "failed to export data")
	}
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(buf.Bytes())
}

func (h *DataHandler) Import(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}
	f, err := file.Open()
	if err != nil {
		return response.InternalError(c, "failed to open file")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 50*1024*1024+1))
	if err != nil {
		return response.InternalError(c, "failed to read file")
	}
	if len(data) > 50*1024*1024 {
		return response.BadRequest(c, "backup exceeds 50 MiB limit")
	}
	result, err := h.svc.Import(data, c.Query("mode", "merge"))
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "import completed", "result": result, "imported": result.Imported, "skipped": result.Skipped, "errors": result.Errors})
}

func (h *DataHandler) DownloadBackup(c *fiber.Ctx) error {
	path, err := h.svc.BackupPath(c.Params("id"))
	if err != nil {
		return response.NotFound(c, "backup")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return response.InternalError(c, "failed to read backup")
	}
	c.Set(fiber.HeaderContentType, "application/zip")
	c.Set(fiber.HeaderContentDisposition, "attachment; filename=\""+filepath.Base(path)+"\"")
	return c.Send(data)
}
