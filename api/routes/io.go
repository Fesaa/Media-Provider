package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ioRoutes struct {
}

func RegisterIoRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	ior := ioRoutes{}
	io := router.Group("/io", auth.Middleware)
	io.Post("/ls", wrap(ior.ListDirs))
	io.Post("/create", wrap(ior.CreateDir))
}

func (ior *ioRoutes) ListDirs(l *log.Logger, ctx *fiber.Ctx) error {
	var req payload.ListDirsRequest
	if err := ctx.BodyParser(&req); err != nil {
		l.Warn("error while parsing query params:", "err", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// TODO: I don't know if this is enough, would need to properly check.
	// This endpoint is behind auth, so you'd already need om some access.
	// But would still want to check.
	cleanedPath := func(p string) string {
		parts := strings.Split(p, "/")
		filtered := slices.DeleteFunc(parts, func(s string) bool {
			return s == ".." || s == "."
		})
		return strings.Join(filtered, "/")
	}(req.Dir)

	root := config.I().GetRootDir()
	entries, err := os.ReadDir(path.Join(root, cleanedPath))
	if err != nil {
		l.Error("error while reading dir:", "err", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var out payload.ListDirResponse
	for _, entry := range entries {
		if !entry.IsDir() && !req.ShowFiles {
			continue
		}
		out = append(out, payload.DirEntry{Name: entry.Name(), Dir: entry.IsDir()})
	}

	return ctx.JSON(out)
}

type CreateDirRequest struct {
	BaseDir string `json:"baseDir"`
	NewDir  string `json:"newDir"`
}

func (ior *ioRoutes) CreateDir(l *log.Logger, ctx *fiber.Ctx) error {
	var req CreateDirRequest
	if err := ctx.BodyParser(&req); err != nil {
		l.Warn("error while parsing query params:", "err", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if strings.Contains(req.NewDir, "..") || strings.Contains(req.BaseDir, "..") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid path",
		})
	}

	root := config.I().GetRootDir()
	p := path.Join(root, req.BaseDir, req.NewDir)
	err := os.Mkdir(p, 0755)
	if err != nil {
		l.Error("error while creating dir:", "err", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusCreated)
}
