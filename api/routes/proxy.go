package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/api/auth"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"io"
	"mime"
	"net/http"
	"path/filepath"
)

type proxyRoutes struct {
	dig.In

	Router     fiber.Router
	Auth       auth.Provider `name:"api-key-auth"`
	Cache      fiber.Handler `name:"cache"`
	Log        zerolog.Logger
	HttpClient *menou.Client
	Transloco  services.TranslocoService
}

func RegisterProxyRoutes(pr proxyRoutes) {
	proxy := pr.Router.Group("/proxy", pr.Cache)
	proxy.Get("/mangadex/covers/:id/:filename", pr.Auth.Middleware, pr.MangaDexCoverProxy)
	proxy.Get("/webtoon/covers/:date/:id/:filename", pr.Auth.Middleware, pr.WebToonCoverProxy)

}

func (pr *proxyRoutes) mangadexUrl(id, fileName string) string {
	return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", id, fileName)
}

func (pr *proxyRoutes) webToonUrl(date, id, fileName string) string {
	return fmt.Sprintf("%s%s/%s/%s?type=q90", webtoon.ImagePrefix, date, id, fileName)
}

func (pr *proxyRoutes) encoding(fileName string) string {
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return mimeType
}

func (pr *proxyRoutes) WebToonCoverProxy(c *fiber.Ctx) error {
	date := c.Params("date")
	id := c.Params("id")
	fileName := c.Params("filename")

	if date == "" || id == "" || fileName == "" {
		return fiber.ErrBadRequest
	}

	url := pr.webToonUrl(date, id, fileName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		pr.Log.Error().Err(err).
			Str("url", url).
			Msg("Failed to create request")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": pr.Transloco.GetTranslation("request-failed"),
		})
	}

	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")

	resp, err := pr.HttpClient.Do(req)
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to send request")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": pr.Transloco.GetTranslation("request-failed"),
		})
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			pr.Log.Warn().Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to download cover image from webtoon")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	c.Set("Content-Type", pr.encoding(fileName))
	return c.Send(data)
}

func (pr *proxyRoutes) MangaDexCoverProxy(c *fiber.Ctx) error {
	id := c.Params("id")
	fileName := c.Params("filename")

	if id == "" || fileName == "" {
		return fiber.ErrBadRequest
	}

	resp, err := pr.HttpClient.Get(pr.mangadexUrl(id, fileName))
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to download cover image from mangadex")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": pr.Transloco.GetTranslation("request-failed"),
		})
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			pr.Log.Warn().Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to download cover image from mangadex")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	c.Set("Content-Type", pr.encoding(fileName))
	return c.Send(data)
}
