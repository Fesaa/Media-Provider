package dynasty

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func NewManga(scope *dig.Scope) api.Downloadable {
	var m *manga

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, client api.Client, httpClient *http.Client,
		log zerolog.Logger, repository Repository, markdownService services.MarkdownService,
	) {
		m = &manga{
			id:              req.Id,
			httpClient:      httpClient,
			repository:      repository,
			markdownService: markdownService,
		}

		d := api.NewDownloadableFromBlock[Chapter](req, m, client, log.With().Str("handler", "dynasty-manga").Logger())
		m.DownloadBase = d
	}))

	return m
}

type manga struct {
	*api.DownloadBase[Chapter]

	httpClient      *http.Client
	repository      Repository
	markdownService services.MarkdownService

	id         string
	seriesInfo *Series
}

func (m *manga) Title() string {
	if m.seriesInfo != nil {
		return m.seriesInfo.Title
	}

	return m.id
}

func (m *manga) Provider() models.Provider {
	return models.DYNASTY
}

func (m *manga) LoadInfo() chan struct{} {
	out := make(chan struct{})
	go func() {
		info, err := m.repository.SeriesInfo(m.id)
		if err != nil {
			m.Log.Error().Err(err).Msg("error while loading series info")
			m.Cancel()
			close(out)
			return
		}

		m.seriesInfo = info
		close(out)
	}()

	return out
}

func (m *manga) GetInfo() payload.InfoStat {
	volumeDiff := m.ImagesDownloaded - m.LastRead
	timeDiff := max(time.Since(m.LastTime).Seconds(), 1)
	speed := max(int64(float64(volumeDiff)/timeDiff), 1)
	m.LastRead = m.ImagesDownloaded
	m.LastTime = time.Now()

	return payload.InfoStat{
		Provider: models.DYNASTY,
		Id:       m.id,
		Name: func() string {
			title := m.Title()
			if title == m.id && m.TempTitle != "" {
				return m.TempTitle
			}
			return title
		}(),
		RefUrl: func() string {
			if m.seriesInfo == nil {
				return ""
			}
			return m.seriesInfo.RefUrl()
		}(),
		Size:        strconv.Itoa(len(m.ToDownload)) + " Chapters",
		Downloading: m.Wg != nil,
		Progress:    utils.Percent(int64(m.ContentDownloaded), int64(len(m.ToDownload))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: m.GetDownloadDir(),
	}
}

func (m *manga) All() []Chapter {
	return m.seriesInfo.Chapters
}

func (m *manga) ContentDir(chapter Chapter) string {
	if chapter.Chapter == "" {
		return fmt.Sprintf("%s OneShot %s", m.Title(), chapter.Title)
	}

	if chpt, err := strconv.ParseFloat(chapter.Chapter, 32); err == nil {
		chDir := fmt.Sprintf("%s Ch. %s", m.Title(), utils.PadFloat(chpt, 4))
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

func (m *manga) ContentUrls(chapter Chapter) ([]string, error) {
	return m.repository.ChapterImages(chapter.Id)
}

func (m *manga) WriteContentMetaData(chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover.jpg")
	if err := m.downloadAndWrite(m.seriesInfo.CoverUrl, filePath); err != nil {
		return err
	}

	m.Log.Trace().Str("chapter", chapter.Chapter).Msg("writing comicinfoxml")
	return comicinfo.Save(m.comicInfo(chapter), path.Join(m.ContentPath(chapter), "ComicInfo.xml"))
}

func (m *manga) comicInfo(chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = m.seriesInfo.Title
	ci.AlternateSeries = m.seriesInfo.AltTitle
	ci.Summary = m.markdownService.SanitizeHtml(m.seriesInfo.Description)
	ci.Manga = comicinfo.MangaYes
	ci.Title = chapter.Title
	if vol, err := strconv.Atoi(chapter.Volume); err == nil {
		ci.Volume = vol
	} else {
		m.Log.Trace().Err(err).Str("chapter", chapter.Volume).Msg("could not convert volume to int")
	}

	ci.Writer = strings.Join(utils.Map(m.seriesInfo.Authors, func(t Author) string {
		return t.DisplayName
	}), ",")

	ci.Tags = strings.Join(utils.Map(utils.FlatMapMany(chapter.Tags, m.seriesInfo.Tags), func(t Tag) string {
		return t.DisplayName
	}), ",")

	ci.Web = m.seriesInfo.RefUrl()

	return ci
}

func (m *manga) DownloadContent(idx int, chapter Chapter, url string) error {
	filePath := path.Join(m.ContentPath(chapter), fmt.Sprintf("page %s.jpg", utils.PadInt(idx, 4)))
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
	if ok || (chapter.Chapter == "" && !m.Req.GetBool(DownloadOneShotKey)) {
		return false
	}

	return true
}

func (m *manga) downloadAndWrite(url string, path string, tryAgain ...bool) error {
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusTooManyRequests {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		if len(tryAgain) > 0 && !tryAgain[0] {
			return fmt.Errorf("hit rate limit too many times")
		}

		d := time.Minute
		m.Log.Warn().Dur("sleeping_for", d).Msg("Hit rate limit, sleeping for 1 minute")
		time.Sleep(d)
		return m.downloadAndWrite(url, path, false)

	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			m.Log.Warn().Err(err).Msg("error closing body")
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, data, 0755); err != nil {
		return err
	}

	return nil
}
