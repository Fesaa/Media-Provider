package routes

import (
	"errors"
	"log/slog"
	"os"
	"path"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

type ListDirsRequest struct {
	Dir string `query:"dir"`
}

func ListDirs(ctx *fiber.Ctx) error {
	var req ListDirsRequest
	if err := ctx.QueryParser(&req); err != nil {
		slog.Error("Error parsing query params:", "err", err)
		return fiber.ErrBadRequest
	}

	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not present while handling login")
		return errors.New("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	tp := holder.GetTorrentProvider()
	if tp == nil {
		slog.Error("No TorrentProvider found while handling login")
		return errors.New("Internal Server Error. \nNo TorrentProvider found. Please contact the administrator.")
	}

	entries, err := os.ReadDir(tp.GetBaseDir() + "/" + req.Dir)
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
		slog.Error("Error parsing query params:", "err", err)
		return fiber.ErrBadRequest
	}

	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not present while handling login")
		return errors.New("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	tp := holder.GetTorrentProvider()
	if tp == nil {
		slog.Error("No TorrentProvider found while handling login")
		return errors.New("Internal Server Error. \nNo TorrentProvider found. Please contact the administrator.")
	}

	path := path.Join(tp.GetBaseDir(), req.BaseDir, req.NewDir)
	err := os.Mkdir(path, 0755)
	if err != nil {
		slog.Error("Error creating dir:", "err", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusCreated)
}
