package dynasty

import (
	"context"
	"errors"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/common"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/spf13/afero"
	"go.uber.org/dig"
)

func New(scope *dig.Scope) (core.Downloadable, error) {
	var m *manga

	err := scope.Invoke(func(
		req payload.DownloadRequest, repository Repository,
		markdownService services.MarkdownService, imageService services.ImageService,
		fs afero.Afero,
	) error {
		m = &manga{
			repository:      repository,
			markdownService: markdownService,
			imageService:    imageService,
			fs:              fs,
		}

		var err error
		m.Core, err = core.New[Chapter, *Series](scope, "dynasty-manga", m)
		return err
	})

	return m, err
}

type manga struct {
	*core.Core[Chapter, *Series]

	repository      Repository
	markdownService services.MarkdownService
	transLoco       services.TranslocoService
	imageService    services.ImageService
	fs              afero.Afero

	coverBytes      []byte
	hasCheckedCover bool
}

func (m *manga) CoreExt() core.Ext[Chapter, *Series] {
	return common.CbzExt[Chapter, *Series]()
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

func (m *manga) Provider() models.Provider {
	return models.DYNASTY
}

func (m *manga) RefUrl() string {
	if m.SeriesInfo == nil {
		return ""
	}

	return m.SeriesInfo.RefUrl()
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

		m.SeriesInfo = info
	}()

	return out
}

func (m *manga) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return m.repository.ChapterImages(ctx, chapter.Id)
}
