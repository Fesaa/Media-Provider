package publication

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type ioTaskFunc func(*publication, context.Context, zerolog.Logger, ioTask) error

type cleanupFunc func(*publication, string) error

// isContentFunc return true if a file with given name should be regarded as downloaded content
// The returned content does not need to be complete, only Content.Volume and Content.Chapter should be
// set if applicable
type isContentFunc func(string) (Content, bool)

type volumeFunc func(*publication, Content) (string, error)

type Extensions struct {
	ioTaskFunc         ioTaskFunc
	contentCleanupFunc cleanupFunc
	isContentFunc      isContentFunc
	volumeFunc         volumeFunc
}

func CbzExt() Extensions {
	return Extensions{
		ioTaskFunc:         imageIoTask,
		contentCleanupFunc: cbzCleanup,
		isContentFunc:      isCbz,
		volumeFunc:         getVolumeFromComicInfo,
	}
}

func imageIoTask(p *publication, ctx context.Context, log zerolog.Logger, task ioTask) error {
	data := task.Data
	ok := false

	pref, err := p.unitOfWork.Preferences.GetPreferences(ctx, p.req.OwnerId)
	if err == nil && pref.ConvertToWebp {
		data, ok = p.imageService.ConvertToWebp(ctx, task.Data)
	}

	ext := utils.Ternary(ok, ".webp", utils.Ext(task.Task.Url.Url))
	filePath := path.Join(task.Path, fmt.Sprintf("page %s"+ext, utils.PadInt(task.Task.Idx, 4)))

	select {
	case <-ctx.Done():
		return nil
	default:
	}

	err = p.fs.WriteFile(filePath, data, 0755)
	if err == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		log.Debug().Err(err).Msg("ignoring write error due to cancellation")
		return nil
	default:
	}
	log.Error().Err(err).Msg("error writing file")
	return err
}

var (
	contentVolumeAndChapterRegex = regexp.MustCompile(".* (?:Vol\\. ([\\d\\.]+)) (?:Ch)\\. ([\\d\\.]+).cbz")
	contentChapterRegex          = regexp.MustCompile(".* Ch\\. ([\\d\\.]+).cbz")
	contentVolumeRegex           = regexp.MustCompile(".* Vol\\. ([\\d\\.]+).cbz")
)

func isCbz(name string) (Content, bool) {
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

func getVolumeFromComicInfo(p *publication, content Content) (string, error) {
	fullPath := path.Join(p.client.GetBaseDir(), content.Path)
	ci, err := p.archiveService.GetComicInfo(fullPath)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(ci.Volume), nil
}

func cbzCleanup(p *publication, path string) error {
	if err := p.dirService.ZipToCbz(path); err != nil {
		return err
	}

	return p.fs.RemoveAll(path)
}
