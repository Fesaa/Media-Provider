package bato

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/spf13/afero"
	"go.uber.org/dig"
)

func NewManga(scope *dig.Scope) core.Downloadable {
	var m *manga

	utils.Must(scope.Invoke(func(req payload.DownloadRequest, repository Repository, fs afero.Afero) {
		m = &manga{
			id:         req.Id,
			repository: repository,
			fs:         fs,
		}

		m.Core = core.New[Chapter](scope, "bato", m)
	}))

	return m
}

type manga struct {
	*core.Core[Chapter]

	repository Repository
	fs         afero.Afero

	id         string
	seriesInfo *Series
}

func (m *manga) RefUrl() string {
	return fmt.Sprintf("%s/title/%s", Domain, m.id)
}

func (m *manga) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})

	go func() {
		defer close(out)
		info, err := m.repository.SeriesInfo(ctx, m.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading series info")
			}
			m.Cancel()
			return
		}

		m.seriesInfo = info
	}()

	return out
}

func (m *manga) All() []Chapter {
	return m.seriesInfo.Chapters
}

func (m *manga) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return m.repository.ChapterImages(ctx, chapter.Id)
}

func (m *manga) ShouldDownload(chapter Chapter) bool {
	_, ok := m.GetContentByName(m.ContentDir(chapter) + ".cbz")
	if ok || (chapter.Chapter == "" && !m.Req.GetBool(core.DownloadOneShotKey)) {
		return false
	}

	return true
}

func (m *manga) Title() string {
	if m.seriesInfo == nil {
		return m.id
	}

	return m.seriesInfo.Title
}

func (m *manga) Provider() models.Provider {
	return models.BATO
}
