package routes

import (
	"github.com/Fesaa/Media-Provider/yoitsu"
	"log/slog"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
)

type ListDirsRequest struct {
	Dir string `json:"dir"`
}

func ListDirs(ctx *fiber.Ctx) error {
	var req ListDirsRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.Warn("Error parsing query params:", "err", err)
		return fiber.ErrBadRequest
	}
	entries, err := os.ReadDir(yoitsu.I().GetBaseDir() + "/" + req.Dir)
	if err != nil {
		slog.Error("Error reading dir:", "err", err)
		return fiber.ErrInternalServerError
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return ctx.JSON(dirs)
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

	p := path.Join(yoitsu.I().GetBaseDir(), req.BaseDir, req.NewDir)
	err := os.Mkdir(p, 0755)
	if err != nil {
		slog.Error("Error creating dir:", "err", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusCreated)
}
