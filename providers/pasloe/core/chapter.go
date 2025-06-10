package core

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"path"
	"regexp"
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

func (c *Core[C, S]) ContentKey(chapter C) string {
	return chapter.GetId()
}

func (c *Core[C, S]) ContentLogger(chapter C) zerolog.Logger {
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
func (c *Core[C, S]) DownloadContent(idx int, chapter C, url string) error {
	filePath := path.Join(c.ContentPath(chapter), fmt.Sprintf("page %s"+utils.Ext(url), utils.PadInt(idx, 4)))
	if err := c.DownloadAndWrite(url, filePath); err != nil {
		return err
	}
	c.ImagesDownloaded++
	return nil
}

func (c *Core[C, S]) ContentPath(chapter C) string {
	base := path.Join(c.Client.GetBaseDir(), c.GetBaseDir(), c.Title())

	if chapter.GetVolume() != "" {
		base = path.Join(base, c.VolumeDir(chapter))
	}

	return path.Join(base, c.ContentDir(chapter))
}

func (c *Core[C, S]) VolumeDir(chapter C) string {
	return fmt.Sprintf("%s Vol. %s", c.Title(), chapter.GetVolume())
}

func (c *Core[C, S]) ContentDir(chapter C) string {
	if chapter.GetChapter() == "" {
		return fmt.Sprintf("%s %s (OneShot)", c.Title(), chapter.GetTitle())
	}

	if _, err := strconv.ParseFloat(chapter.GetChapter(), 32); err == nil {
		padded := utils.PadFloatFromString(chapter.GetChapter(), 4)
		return fmt.Sprintf("%s Ch. %s", c.Title(), padded)
	} else if chapter.GetChapter() != "" {
		c.Log.Warn().Err(err).Str("chapter", chapter.GetChapter()).Msg("unable to parse chapter number, not padding")
	}

	return fmt.Sprintf("%s Ch. %s", c.Title(), chapter.GetChapter())
}

var (
	contentRegex    = regexp.MustCompile(".* (?:Ch|Vol)\\. ([\\d|\\.]+).cbz")
	oneShotRegexOld = regexp.MustCompile(".+ OneShot .+\\.cbz")
	oneShotRegex    = regexp.MustCompile(".+ \\(OneShot\\).cbz")
)

func (c *Core[C, S]) IsContent(name string) bool {
	if contentRegex.MatchString(name) {
		return true
	}

	if oneShotRegex.MatchString(name) {
		return true
	}

	if oneShotRegexOld.MatchString(name) {
		return true
	}

	return false
}
