package core

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"path"
	"strconv"
)

// Chapter represents any downloadable group of images
//
// A chapter is considered standalone/OneShot if GetVolume and GetChapter both return an empty string
type Chapter interface {
	GetId() string
	Label() string

	GetChapter() string
	GetVolume() string
	GetTitle() string
}

func (c *Core[T]) ContentKey(chapter T) string {
	return chapter.GetId()
}

func (c *Core[T]) ContentLogger(chapter T) zerolog.Logger {
	builder := c.Log.With().
		Str("chapterId", chapter.GetId()).
		Str("chapter", chapter.GetChapter())

	if chapter.GetTitle() != "" {
		builder = builder.Str("title", chapter.GetTitle())
	}

	if chapter.GetVolume() != "" {
		builder = builder.Str("volume", chapter.GetVolume())
	}

	return builder.Logger()
}

// DownloadContent TODO: Add context.Context
func (c *Core[T]) DownloadContent(idx int, chapter T, url string) error {
	filePath := path.Join(c.ContentPath(chapter), fmt.Sprintf("page %s"+utils.Ext(url), utils.PadInt(idx, 4)))
	if err := c.DownloadAndWrite(url, filePath); err != nil {
		return err
	}
	c.ImagesDownloaded++
	return nil
}

func (c *Core[T]) ContentPath(chapter T) string {
	base := path.Join(c.Client.GetBaseDir(), c.GetBaseDir(), c.infoProvider.Title())

	if chapter.GetVolume() != "" {
		base = path.Join(base, fmt.Sprintf("%s Vol. %s", c.infoProvider.Title(), chapter.GetVolume()))
	}

	return path.Join(base, c.ContentDir(chapter))
}

func (c *Core[T]) ContentDir(chapter T) string {
	if chapter.GetChapter() == "" {
		return fmt.Sprintf("%s %s (OneShot)", c.infoProvider.Title(), chapter.GetTitle())
	}

	if _, err := strconv.ParseFloat(chapter.GetChapter(), 32); err == nil {
		padded := utils.PadFloatFromString(chapter.GetChapter(), 4)
		return fmt.Sprintf("%s Ch. %s", c.infoProvider.Title(), padded)
	} else if chapter.GetChapter() != "" {
		c.Log.Warn().Err(err).Str("chapter", chapter.GetChapter()).Msg("unable to parse chapter number, not padding")
	}

	return fmt.Sprintf("%s Ch. %s", c.infoProvider.Title(), chapter.GetChapter())
}

func IsOneShot(chapter Chapter) bool {
	return chapter.GetChapter() == "" && chapter.GetVolume() == ""
}
