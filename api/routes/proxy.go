package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/wisewolf"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime"
	"net/http"
	"path/filepath"
)

type proxyRoutes struct {
}

func RegisterProxyRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	pr := proxyRoutes{}
	proxy := router.Group("/proxy", cache)
	proxy.Get("/mangadex/covers/:id/:filename", auth.MiddlewareWithApiKey, wrap(pr.MangaDexCoverProxy))
	proxy.Get("/webtoon/covers/:date/:id/:filename", auth.MiddlewareWithApiKey, wrap(pr.WebToonCoverProxy))

}

func (pr *proxyRoutes) mangadexUrl(id, fileName string) string {
	return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", id, fileName)
}

func (pr *proxyRoutes) webToonUrl(date, id, fileName string) string {
	return fmt.Sprintf("https://webtoon-phinf.pstatic.net/%s/%s/%s?type=q90", date, id, fileName)
}

func (pr *proxyRoutes) encoding(fileName string) string {
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return mimeType
}

func (pr *proxyRoutes) WebToonCoverProxy(l *log.Logger, c *fiber.Ctx) error {
	date := c.Params("date")
	id := c.Params("id")
	fileName := c.Params("filename")

	if date == "" || id == "" || fileName == "" {
		return fiber.ErrBadRequest
	}

	req, err := http.NewRequest(http.MethodGet, pr.webToonUrl(date, id, fileName), nil)
	if err != nil {
		l.Error("Failed to construct new request", "error", err)
		return fiber.ErrInternalServerError
	}

	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")

	resp, err := wisewolf.Client.Do(req)
	if err != nil {
		l.Error("Failed to make request", "error", err)
		return fiber.ErrInternalServerError
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			l.Warn("Failed to close response body", "error", err)
		}
	}(resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to download cover image from webtoon", "error", err)
		return fiber.ErrInternalServerError
	}

	c.Set("Content-Type", pr.encoding(fileName))
	return c.Send(data)
}

func (pr *proxyRoutes) MangaDexCoverProxy(l *log.Logger, c *fiber.Ctx) error {
	id := c.Params("id")
	fileName := c.Params("filename")

	if id == "" || fileName == "" {
		return fiber.ErrBadRequest
	}

	resp, err := wisewolf.Client.Get(pr.mangadexUrl(id, fileName))
	if err != nil {
		l.Error("Failed to download cover image from mangadex", "error", err)
		return fiber.ErrInternalServerError
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			l.Warn("Failed to close response body", "error", err)
		}
	}(resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to download cover image from mangadex", "error", err)
		return fiber.ErrInternalServerError
	}

	c.Set("Content-Type", pr.encoding(fileName))
	return c.Send(data)
}
