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

func New(scope *dig.Scope) core.Downloadable {
	var m *manga

	utils.Must(scope.Invoke(func(req payload.DownloadRequest, repository Repository, fs afero.Afero) {
		m = &manga{
			repository: repository,
			fs:         fs,
		}

		m.Core = core.New[Chapter, *Series](scope, "bato", m)
	}))

	return m
}

type manga struct {
	*core.Core[Chapter, *Series]

	repository Repository
	fs         afero.Afero
}

func (m *manga) Title() string {
	if titleOverride, ok := m.Req.GetString(core.TitleOverride); ok {
		return titleOverride
	}

	if m.SeriesInfo == nil {
		return utils.NonEmpty(m.Req.TempTitle, m.Req.Id)
	}

	return utils.NonEmpty(m.SeriesInfo.GetTitle(), m.Req.TempTitle, m.Req.Id)
}

func (m *manga) RefUrl() string {
	return fmt.Sprintf("%s/title/%s", Domain, m.Id())
}

func (m *manga) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})

	go func() {
		defer close(out)
		info, err := m.repository.SeriesInfo(ctx, m.Id())
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading series info")
			}
			m.Cancel()
			return
		}

		m.SeriesInfo = &info
	}()

	return out
}

func (m *manga) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return m.repository.ChapterImages(ctx, chapter.Id)
}

func (m *manga) Provider() models.Provider {
	return models.BATO
}
