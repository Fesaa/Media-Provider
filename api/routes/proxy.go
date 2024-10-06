package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime"
	"net/http"
	"path/filepath"
)

func url(id, fileName string) string {
	return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", id, fileName)
}

func encoding(fileName string) string {
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return mimeType
}

func MangaDexCoverProxy(c *fiber.Ctx) error {
	id := c.Params("id")
	fileName := c.Params("filename")

	if id == "" || fileName == "" {
		return fiber.ErrBadRequest
	}

	resp, err := http.Get(url(id, fileName))
	if err != nil {
		log.Error("Failed to download cover image from mangadex", "error", err)
		return fiber.ErrInternalServerError
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Warn("Failed to close response body", "error", err)
		}
	}(resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to download cover image from mangadex", "error", err)
		return fiber.ErrInternalServerError
	}

	c.Set("Content-Type", encoding(fileName))
	return c.Send(data)
}
