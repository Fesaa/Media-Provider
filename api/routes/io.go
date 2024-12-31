package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ioRoutes struct {
	dig.In

	Router fiber.Router
	Cfg    *config.Config
	Auth   auth.Provider `name:"jwt-auth"`
	Log    zerolog.Logger
}

func RegisterIoRoutes(ior ioRoutes) {
	io := ior.Router.Group("/io", ior.Auth.Middleware)
	io.Post("/ls", ior.ListDirs)
	io.Post("/create", ior.CreateDir)
}

func (ior *ioRoutes) ListDirs(ctx *fiber.Ctx) error {
	var req payload.ListDirsRequest
	if err := ctx.BodyParser(&req); err != nil {
		ior.Log.Warn().Err(err).Msg("failed to parse request")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// TODO: I don't know if this is enough, would need to properly check.
	// This endpoint is behind auth, so you'd already need om some access.
	// But would still want to check.
	cleanedPath := CleanPath(req.Dir)

	root := ior.Cfg.GetRootDir()
	entries, err := os.ReadDir(path.Join(root, cleanedPath))
	if err != nil {
		ior.Log.Warn().Err(err).Msg("failed to read dir")
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

func (ior *ioRoutes) CreateDir(ctx *fiber.Ctx) error {
	var req CreateDirRequest
	if err := ctx.BodyParser(&req); err != nil {
		ior.Log.Warn().Err(err).Msg("failed to parse request")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if strings.Contains(req.NewDir, "..") || strings.Contains(req.BaseDir, "..") {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid path",
		})
	}

	root := ior.Cfg.GetRootDir()
	p := path.Join(root, req.BaseDir, req.NewDir)
	err := os.Mkdir(p, 0755)
	if err != nil {
		ior.Log.Warn().Err(err).Msg("failed to create dir")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

func CleanPath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")
	filtered := slices.DeleteFunc(parts, func(s string) bool {
		return s == ".." || s == "."
	})
	return strings.Join(filtered, "/")
}
