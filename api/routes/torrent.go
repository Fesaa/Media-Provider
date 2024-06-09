package routes

import (
	"fmt"
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

type DownloadRequest struct {
	Info    string `json:"info"`
	BaseDir string `json:"base_dir"`
	Url     bool   `json:"url"`
}

func (d DownloadRequest) DebugString() string {
	return fmt.Sprintf("{Info: %s, BaseDir: %s, Url: %t}", d.Info, d.BaseDir, d.Url)
}

func Download(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Debug("Holder not found in context")
		return fiber.ErrInternalServerError
	}

	torrentProvider := holder.GetTorrentProvider()
	if torrentProvider == nil {
		slog.Debug("torrentImpl provider not found in holder")
		return fiber.ErrInternalServerError
	}

	var req DownloadRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		slog.Error("Error parsing request body into DownloadRequest", "err", err)
		return fiber.ErrBadRequest
	}

	if req.BaseDir == "" {
		slog.Warn("Trying to download Torrent to empty baseDir, returning error.")
		return fiber.ErrBadRequest
	}

	if req.Url {
		_, err = torrentProvider.AddDownloadFromUrl(req.Info, req.BaseDir)
	} else {
		_, err = torrentProvider.AddDownload(req.Info, req.BaseDir)
	}

	if err != nil {
		slog.Error("Error adding download", "error", err, "debug_info", req.DebugString())
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stop(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Debug("Holder not found in context")
		return fiber.ErrInternalServerError
	}

	torrentProvider := holder.GetTorrentProvider()
	if torrentProvider == nil {
		slog.Debug("torrentImpl provider not found in holder")
		return fiber.ErrInternalServerError
	}

	infoHash := ctx.Params("infoHash")
	if infoHash == "" {
		slog.Error("No infoHash provided")
		return fiber.ErrBadRequest
	}

	err := torrentProvider.RemoveDownload(infoHash, true)
	if err != nil {
		slog.Error("Error stopping download", "infoHash", infoHash, "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stats(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not found in context")
		return fiber.ErrInternalServerError
	}

	torrentProvider := holder.GetTorrentProvider()
	if torrentProvider == nil {
		slog.Error("torrentImpl provider not found in holder")
		return fiber.ErrInternalServerError
	}

	torrents := torrentProvider.GetRunningTorrents()
	info := make(map[string]models.TorrentInfo, torrents.Len())
	torrents.ForEachSafe(func(key string, torrent models.Torrent) {
		info[key] = torrent.GetInfo()
	})
	return ctx.JSON(info)
}
