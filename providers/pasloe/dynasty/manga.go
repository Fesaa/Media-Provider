package dynasty

import (
	"context"
	"errors"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/spf13/afero"
	"go.uber.org/dig"
)

func NewManga(scope *dig.Scope) core.Downloadable {
	var m *manga

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, httpClient *menou.Client,
		repository Repository, markdownService services.MarkdownService,
		preferences models.Preferences, imageService services.ImageService,
		fs afero.Afero,
	) {
		m = &manga{
			id:              req.Id,
			httpClient:      httpClient,
			repository:      repository,
			markdownService: markdownService,
			preferences:     preferences,
			imageService:    imageService,
			fs:              fs,
		}

		m.Core = core.New[Chapter](scope, "dynasty-manga", m)
	}))

	return m
}

type manga struct {
	*core.Core[Chapter]

	httpClient      *menou.Client
	repository      Repository
	markdownService services.MarkdownService
	preferences     models.Preferences
	transLoco       services.TranslocoService
	imageService    services.ImageService
	fs              afero.Afero

	id              string
	seriesInfo      *Series
	coverBytes      []byte
	hasCheckedCover bool

	hasWarnedBlacklist bool
}

func (m *manga) Title() string {
	if m.seriesInfo != nil {
		return m.seriesInfo.Title
	}

	if temp := m.Req.TempTitle; temp != "" {
		return temp
	}

	return m.id
}

func (m *manga) Provider() models.Provider {
	return models.DYNASTY
}

func (m *manga) RefUrl() string {
	if m.seriesInfo == nil {
		return ""
	}

	return m.seriesInfo.RefUrl()
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
