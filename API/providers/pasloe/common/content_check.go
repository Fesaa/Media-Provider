package common

import (
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
)

var (
	contentVolumeAndChapterRegex = regexp.MustCompile(".* (?:Vol\\. ([\\d\\.]+)) (?:Ch)\\. ([\\d\\.]+).cbz")
	contentChapterRegex          = regexp.MustCompile(".* Ch\\. ([\\d\\.]+).cbz")
	contentVolumeRegex           = regexp.MustCompile(".* Vol\\. ([\\d\\.]+).cbz")
)

func IsCbz(name string) (core.Content, bool) {
	matches := contentVolumeAndChapterRegex.FindStringSubmatch(name)
	if len(matches) == 3 {
		return core.Content{
			Volume:  utils.TrimLeadingZero(matches[1]),
			Chapter: utils.TrimLeadingZero(matches[2]),
		}, true
	}

	matches = contentVolumeRegex.FindStringSubmatch(name)
	if len(matches) == 2 {
		return core.Content{
			Volume: utils.TrimLeadingZero(matches[1]),
		}, true
	}

	matches = contentChapterRegex.FindStringSubmatch(name)
	if len(matches) == 2 {
		return core.Content{
			Chapter: utils.TrimLeadingZero(matches[1]),
		}, true
	}

	// Fallback to simple ext check
	return core.Content{}, filepath.Ext(name) == ".cbz"
}

func GetVolumeFromComicInfo[C core.Chapter, S core.Series[C]](c *core.Core[C, S], content core.Content) (string, error) {
	fullPath := path.Join(c.Client.GetBaseDir(), content.Path)
	ci, err := c.ArchiveService.GetComicInfo(fullPath)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(ci.Volume), nil
}
