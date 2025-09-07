package routes

import (
	"errors"
	"path"
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/spf13/afero"
	"go.uber.org/dig"

	"github.com/gofiber/fiber/v2"
)

type ioRoutes struct {
	dig.In

	Router          fiber.Router
	Cfg             *config.Config
	Auth            services.AuthService
	Val             services.ValidationService
	Transloco       services.TranslocoService
	SettingsService services.SettingsService
	Fs              afero.Afero
}

func RegisterIoRoutes(ior ioRoutes) {
	ior.Router.Group("/io", ior.Auth.Middleware).
		Post("/ls", withBodyValidation(ior.listDirs)).
		Post("/create", withBodyValidation(ior.createDir))
}

func (ior *ioRoutes) listDirs(ctx *fiber.Ctx, req payload.ListDirsRequest) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	settings, err := ior.SettingsService.GetSettingsDto(ctx.UserContext())
	if err != nil {
		return InternalError(err)
	}

	entries, err := ior.Fs.ReadDir(path.Join(settings.RootDir, path.Clean(req.Dir)))
	if err != nil {
		log.Warn().Err(err).Msg("failed to read dir")
		return InternalError(err)
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
	log := services.GetFromContext(ctx, services.LoggerKey)

	if strings.Contains(req.NewDir, "..") || strings.Contains(req.BaseDir, "..") {
		log.Warn().Str("newDir", req.NewDir).Str("baseDir", req.BaseDir).
			Msg("path contained invalid characters")
		return BadRequest(errors.New(ior.Transloco.GetTranslation("invalid-path")))
	}

	settings, err := ior.SettingsService.GetSettingsDto(ctx.UserContext())
	if err != nil {
		return InternalError(err)
	}

	p := path.Join(settings.RootDir, req.BaseDir, path.Clean(req.NewDir))
	if err = ior.Fs.Mkdir(p, 0755); err != nil {
		log.Warn().Err(err).Msg("failed to create dir")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{})
}
