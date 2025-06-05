package bato

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"io"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strconv"
	"time"
)

func NewManga(scope *dig.Scope) api.Downloadable {
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

		m.DownloadBase = api.NewBaseWithProvider[Chapter](scope, "dynasty-manga", m)
	}))

	return m
}

type manga struct {
	*api.DownloadBase[Chapter]

	httpClient      *menou.Client
	repository      Repository
	markdownService services.MarkdownService
	preferences     models.Preferences
	transLoco       services.TranslocoService
	imageService    services.ImageService
	fs              afero.Afero

	id                 string
	seriesInfo         *Series
	hasWarnedBlacklist bool
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

func (m *manga) ContentList() []payload.ListContentData {
	if m.seriesInfo == nil {
		return nil
	}

	data := utils.GroupBy(m.seriesInfo.Chapters, func(v Chapter) string {
		return v.Volume
	})

	childrenFunc := func(chapters []Chapter) []payload.ListContentData {
		slices.SortFunc(chapters, func(a, b Chapter) int {
			if a.Volume != b.Volume {
				return (int)(b.VolumeFloat() - a.VolumeFloat())
			}
			return (int)(b.ChapterFloat() - a.ChapterFloat())
		})

		return utils.Map(chapters, func(chapter Chapter) payload.ListContentData {
			return payload.ListContentData{
				SubContentId: chapter.Id,
				Selected:     len(m.ToDownloadUserSelected) == 0 || slices.Contains(m.ToDownloadUserSelected, chapter.Id),
				Label:        utils.Ternary(chapter.Title == "", m.Title()+" "+chapter.Label(), chapter.Label()),
			}
		})
	}

	sortSlice := utils.Keys(data)
	slices.SortFunc(sortSlice, utils.SortFloats)

	var out []payload.ListContentData
	for _, volume := range sortSlice {
		chapters := data[volume]

		// Do not add No Volume label if there are no volumes
		if volume == "" && len(sortSlice) == 1 {
			out = append(out, childrenFunc(chapters)...)
			continue
		}

		out = append(out, payload.ListContentData{
			Label:    utils.Ternary(volume == "", "No Volume", fmt.Sprintf("Volume %s", volume)),
			Children: childrenFunc(chapters),
		})
	}
	return out
}

func (m *manga) ContentDir(chapter Chapter) string {
	if chapter.Chapter == "" {
		return fmt.Sprintf("%s OneShot %s", m.Title(), chapter.Title)
	}

	if _, err := strconv.ParseFloat(chapter.Chapter, 32); err == nil {
		padded := utils.PadFloatFromString(chapter.Chapter, 4)
		chDir := fmt.Sprintf("%s Ch. %s", m.Title(), padded)
		return chDir
	} else if chapter.Chapter != "" {
		m.Log.Warn().Err(err).Str("chapter", chapter.Chapter).Msg("unable to parse chapter number, not padding")
	}

	return fmt.Sprintf("%s Ch. %s", m.Title(), chapter.Chapter)
}

func (m *manga) ContentPath(chapter Chapter) string {
	return path.Join(m.Client.GetBaseDir(), m.GetBaseDir(), m.Title(), m.ContentDir(chapter))
}

func (m *manga) ContentKey(chapter Chapter) string {
	return chapter.Id
}

func (m *manga) ContentLogger(chapter Chapter) zerolog.Logger {
	builder := m.Log.With().
		Str("chapterId", chapter.Id).
		Str("chapter", chapter.Chapter)

	if chapter.Title != "" {
		builder = builder.Str("title", chapter.Title)
	}

	if chapter.Volume != "" {
		builder = builder.Str("volume", chapter.Volume)
	}

	return builder.Logger()
}

func (m *manga) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return m.repository.ChapterImages(ctx, chapter.Id)
}

func (m *manga) DownloadContent(idx int, chapter Chapter, url string) error {
	filePath := path.Join(m.ContentPath(chapter), fmt.Sprintf("page %s"+utils.Ext(url), utils.PadInt(idx, 4)))
	if err := m.downloadAndWrite(url, filePath); err != nil {
		return err
	}
	m.ImagesDownloaded++
	return nil
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
	_, ok := m.GetContentByName(m.ContentDir(chapter) + ".cbz")
	if ok || (chapter.Chapter == "" && !m.Req.GetBool(api.DownloadOneShotKey)) {
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

func (m *manga) download(url string, tryAgain ...bool) ([]byte, error) {
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusTooManyRequests {
			return nil, fmt.Errorf("bad status: %s", resp.Status)
		}

		if len(tryAgain) > 0 && !tryAgain[0] {
			return nil, fmt.Errorf("hit rate limit too many times")
		}

		d := time.Minute
		m.Log.Warn().Dur("sleeping_for", d).Msg("Hit rate limit, sleeping for 1 minute")
		time.Sleep(d)
		return m.download(url, false)

	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			m.Log.Warn().Err(err).Msg("error closing body")
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *manga) downloadAndWrite(url string, path string, tryAgain ...bool) error {
	data, err := m.download(url, tryAgain...)
	if err != nil {
		return err
	}

	if err = m.fs.WriteFile(path, data, 0755); err != nil {
		return err
	}

	return nil
}
