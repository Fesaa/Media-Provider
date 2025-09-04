package routes

import (
	"path"
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"

	"github.com/gofiber/fiber/v2"
)

type ioRoutes struct {
	dig.In

	Router          fiber.Router
	Cfg             *config.Config
	Auth            services.AuthService `name:"jwt-auth"`
	Log             zerolog.Logger
	Val             services.ValidationService
	Transloco       services.TranslocoService
	SettingsService services.SettingsService
	Fs              afero.Afero
}

func RegisterIoRoutes(ior ioRoutes) {
	io := ior.Router.Group("/io", ior.Auth.Middleware)
	io.Post("/ls", withBodyValidation(ior.listDirs))
	io.Post("/create", withBodyValidation(ior.createDir))
}

func (ior *ioRoutes) listDirs(ctx *fiber.Ctx, req payload.ListDirsRequest) error {
	settings, err := ior.SettingsService.GetSettingsDto()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	entries, err := ior.Fs.ReadDir(path.Join(settings.RootDir, path.Clean(req.Dir)))
	if err != nil {
		ior.Log.Warn().Err(err).Msg("failed to read dir")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
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

type createDirRequest struct {
	BaseDir string `json:"baseDir"`
	NewDir  string `json:"newDir"`
}

func (ior *ioRoutes) createDir(ctx *fiber.Ctx, req createDirRequest) error {
	if strings.Contains(req.NewDir, "..") || strings.Contains(req.BaseDir, "..") {
		ior.Log.Warn().Str("newDir", req.NewDir).Str("baseDir", req.BaseDir).
			Msg("path contained invalid characters")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": ior.Transloco.GetTranslation("invalid-path"),
		})
	}

	settings, err := ior.SettingsService.GetSettingsDto()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	p := path.Join(settings.RootDir, req.BaseDir, path.Clean(req.NewDir))
	if err = ior.Fs.Mkdir(p, 0755); err != nil {
		ior.Log.Warn().Err(err).Msg("failed to create dir")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusCreated)
}
