package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"log/slog"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
)

type ListDirsRequest struct {
	Dir       string `json:"dir"`
	ShowFiles bool   `json:"files"`
}

type DirEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
}

func ListDirs(ctx *fiber.Ctx) error {
	var req ListDirsRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.Warn("Error parsing query params:", "err", err)
		return fiber.ErrBadRequest
	}

	base := config.OrDefault(config.I().GetRootDir(), "temp")
	entries, err := os.ReadDir(path.Join(base, req.Dir))
	if err != nil {
		slog.Error("Error reading dir:", "err", err)
		return fiber.ErrInternalServerError
	}

	var out []DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && !req.ShowFiles {
			continue
		}
		out = append(out, DirEntry{entry.Name(), entry.IsDir()})
	}

	return ctx.JSON(out)
}

type CreateDirRequest struct {
	BaseDir string `json:"baseDir"`
	NewDir  string `json:"newDir"`
}

func CreateDir(ctx *fiber.Ctx) error {
	var req CreateDirRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.Warn("Error parsing query params:", "err", err)
		return fiber.ErrBadRequest
	}

	base := config.OrDefault(config.I().GetRootDir(), "temp")
	p := path.Join(base, req.BaseDir, req.NewDir)
	err := os.Mkdir(p, 0755)
	if err != nil {
		slog.Error("Error creating dir:", "err", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusCreated)
}
