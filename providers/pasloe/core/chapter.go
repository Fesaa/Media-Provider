package core

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"path"
	"regexp"
	"slices"
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
	data, err := c.Download(url)
	if err != nil {
		return err
	}

	data, ok := c.imageService.ConvertToWebp(data)

	ext := utils.Ternary(ok, ".webp", utils.Ext(url))
	filePath := path.Join(c.ContentPath(chapter), fmt.Sprintf("page %s"+ext, utils.PadInt(idx, 4)))

	if err = c.fs.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	c.ImagesDownloaded++
	return nil
}

func (c *Core[C, S]) ContentPath(chapter C) string {
	base := path.Join(c.Client.GetBaseDir(), c.GetBaseDir(), c.impl.Title())

	if chapter.GetVolume() != "" && !config.DisableVolumeDirs {
		base = path.Join(base, c.VolumeDir(chapter))
	}

	return path.Join(base, c.ContentDir(chapter))
}

func (c *Core[C, S]) VolumeDir(chapter C) string {
	return fmt.Sprintf("%s Vol. %s", c.impl.Title(), chapter.GetVolume())
}

func (c *Core[C, S]) ContentDir(chapter C) string {
	if chapter.GetChapter() == "" {
		oneShotPath := fmt.Sprintf("%s %s", c.impl.Title(), chapter.GetTitle())
		if !config.DisableOneShotInFileName {
			oneShotPath += " (One Shot)"
		}

		finalOneShotPath := oneShotPath
		for i := 0; slices.Contains(c.HasDownloaded, finalOneShotPath); i++ {
			finalOneShotPath = fmt.Sprintf("%s (%d)", oneShotPath, i)
			if i >= 25 {
				log := c.ContentLogger(chapter)
				log.Warn().Int("tries", i).Msg("Amount of unnamed, or same named OneShots has exceeded 25. Falling back to random generated string")
				finalOneShotPath = fmt.Sprintf("%s (%s)", oneShotPath, utils.MustReturn(utils.GenerateSecret(8)))
			}
		}

		return finalOneShotPath
	}

	fileName := c.impl.Title()
	// Add vol marker in the file name when not using volume dirs
	if chapter.GetVolume() != "" && config.DisableVolumeDirs {
		fileName += fmt.Sprintf(" Vol. %s", chapter.GetVolume())
	}

	if _, err := strconv.ParseFloat(chapter.GetChapter(), 32); err == nil {
		padded := utils.PadFloatFromString(chapter.GetChapter(), 4)
		return fmt.Sprintf("%s Ch. %s", fileName, padded)
	} else if chapter.GetChapter() != "" {
		c.Log.Warn().Err(err).Str("chapter", chapter.GetChapter()).Msg("unable to parse chapter number, not padding")
	}

	return fmt.Sprintf("%s Ch. %s", fileName, chapter.GetChapter())
}

var (
	contentRegex    = regexp.MustCompile(".* (?:Vol\\. [\\d|\\.]+ )?(?:Ch|Vol)\\. ([\\d|\\.]+).cbz")
	oneShotRegexOld = regexp.MustCompile(".+ One ?Shot .+\\.cbz")
	oneShotRegex    = regexp.MustCompile(".+ \\(One ?Shot\\).cbz")
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
