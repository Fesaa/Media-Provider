package dynasty

import (
	"context"
	"errors"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"path"
	"regexp"
	"strconv"
)

func (m *manga) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		info, err := m.repository.SeriesInfo(ctx, m.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading series info")
			}
			m.Cancel()
			close(out)
			return
		}

		m.seriesInfo = info
		close(out)
	}()

	return out
}

var (
	chapterRegex = regexp.MustCompile(".* Ch\\. ([\\d|\\.]+).cbz")
	oneShotRegex = regexp.MustCompile(".+ OneShot .+\\.cbz")
)

func (m *manga) IsContent(name string) bool {
	if chapterRegex.MatchString(name) {
		return true
	}

	if oneShotRegex.MatchString(name) {
		return true
	}

	return false
}

func (m *manga) ShouldDownload(chapter Chapter) bool {
	content, ok := m.GetContentByName(m.ContentDir(chapter) + ".cbz")
	if !ok {
		// Content not on disk, download if not a OneShot, or if we want to download OneShots
		return chapter.Chapter != "" || m.Req.GetBool(DownloadOneShotKey)
	}

	if chapter.Chapter == "" && !m.Req.GetBool(DownloadOneShotKey) {
		return false
	}

	// No need for I/O if there is no volume
	if chapter.Volume == "" {
		return false
	}

	return m.replaceAndShouldDownload(chapter, content)
}

func (m *manga) replaceAndShouldDownload(chapter Chapter, content api.Content) bool {
	l := m.ContentLogger(chapter)
	fullPath := path.Join(m.Client.GetBaseDir(), content.Path)

	ci, err := m.metadataService.GetComicInfo(fullPath)
	if err != nil {
		l.Warn().Err(err).Str("path", fullPath).Msg("unable to read comic info in zip")
		return false
	}

	if strconv.Itoa(ci.Volume) == chapter.Volume {
		l.Trace().Str("path", fullPath).Msg("Volume on disk matches, not re-downloading")
		// Dynasty doesn't have nice covers anyway, don't bother checking
		return false
	}

	l.Debug().Int("onDiskVolume", ci.Volume).Str("path", fullPath).
		Msg("Loose chapter has been assigned to a volume, replacing")
	return true
}
