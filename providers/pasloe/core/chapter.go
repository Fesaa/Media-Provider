package core

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
)

// Chapter represents any downloadable group of images
//
// A chapter is considered standalone/OneShot if GetChapter returns an empty string
type Chapter interface {
	GetId() string

	GetChapter() string
	GetVolume() string
	GetTitle() string
}

func ChapterLabel(c Chapter) string {
	if c.GetChapter() != "" && c.GetVolume() != "" {
		return fmt.Sprintf("Volume %s Chapter %s: %s", c.GetChapter(), c.GetVolume(), c.GetTitle())
	}

	if c.GetChapter() != "" {
		return fmt.Sprintf("Chapter %s: %s", c.GetChapter(), c.GetTitle())
	}

	return fmt.Sprintf("OneShot: %s", c.GetTitle())
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

func (c *Core[C, S]) DownloadContent(ctx context.Context, idx int, chapter C, url string) error {
	data, err := c.Download(ctx, url)
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

// ContentPath returns the full path to the directory where images, and metadata for a chapter
// should be downloaded to
func (c *Core[C, S]) ContentPath(chapter C) string {
	base := path.Join(c.Client.GetBaseDir(), c.GetBaseDir(), c.impl.Title())

	if chapter.GetVolume() != "" && !config.DisableVolumeDirs {
		base = path.Join(base, c.VolumeDir(chapter))
	}

	return path.Join(base, c.ContentFileName(chapter))
}

func (c *Core[C, S]) VolumeDir(chapter C) string {
	return fmt.Sprintf("%s Vol. %s", c.impl.Title(), chapter.GetVolume())
}

// ContentFileName returns the final file name for the downloaded Chapter
// This will be used as a directory until the content is zipped
func (c *Core[C, S]) ContentFileName(chapter C) string {
	if chapter.GetChapter() == "" {
		return c.OneShotFileName(chapter)
	}

	return c.DefaultFileName(chapter)
}

func (c *Core[C, S]) DefaultFileName(chapter C) string {
	fileName := c.impl.Title()

	if chapter.GetVolume() != "" && c.ShouldIncludeVolume() {
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

func (c *Core[C, S]) ShouldIncludeVolume() bool {
	if config.DisableVolumeDirs {
		return true
	}

	if b, ok := c.hasDuplicatedChapters.Get(); ok {
		return b
	}

	groupedByChapter := utils.GroupBy(c.GetAllLoadedChapters(), func(v C) string {
		return v.GetChapter()
	})

	for _, chapterGroup := range groupedByChapter {
		if len(chapterGroup) > 1 {
			c.hasDuplicatedChapters.Set(true)
			return true
		}
	}

	c.hasDuplicatedChapters.Set(false)
	return false
}

func (c *Core[C, S]) OneShotFileName(chapter C) string {
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

var (
	contentVolumeAndChapterRegex = regexp.MustCompile(".* (?:Vol\\. ([\\d\\.]+)) (?:Ch)\\. ([\\d\\.]+).cbz")
	contentChapterRegex          = regexp.MustCompile(".* Ch\\. ([\\d\\.]+).cbz")
	contentVolumeRegex           = regexp.MustCompile(".* Vol\\. ([\\d\\.]+).cbz")
)

func (c *Core[C, S]) IsContent(name string) (Content, bool) {
	matches := contentVolumeAndChapterRegex.FindStringSubmatch(name)
	if len(matches) == 3 {
		return Content{
			Volume:  utils.TrimLeadingZero(matches[1]),
			Chapter: utils.TrimLeadingZero(matches[2]),
		}, true
	}

	matches = contentVolumeRegex.FindStringSubmatch(name)
	if len(matches) == 2 {
		return Content{
			Volume: utils.TrimLeadingZero(matches[1]),
		}, true
	}

	matches = contentChapterRegex.FindStringSubmatch(name)
	if len(matches) == 2 {
		return Content{
			Chapter: utils.TrimLeadingZero(matches[1]),
		}, true
	}

	// Fallback to simple ext check
	return Content{}, filepath.Ext(name) == ".cbz"
}
