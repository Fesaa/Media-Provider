package dynasty

import (
	"context"
	"errors"
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
	"slices"
	"strconv"
	"strings"
	"time"
)

func NewManga(scope *dig.Scope) api.Downloadable {
	var m *manga

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, client api.Client, httpClient *http.Client,
		log zerolog.Logger, repository Repository, markdownService services.MarkdownService,
		signalR services.SignalRService, notification services.NotificationService,
		preferences models.Preferences,
	) {
		m = &manga{
			id:              req.Id,
			httpClient:      httpClient,
			repository:      repository,
			markdownService: markdownService,
			preferences:     preferences,
		}

		d := api.NewDownloadableFromBlock[Chapter](req, m, client,
			log.With().Str("handler", "dynasty-manga").Logger(), signalR, notification)
		m.DownloadBase = d
	}))

	return m
}

type manga struct {
	*api.DownloadBase[Chapter]

	httpClient      *http.Client
	repository      Repository
	markdownService services.MarkdownService
	preferences     models.Preferences

	id         string
	seriesInfo *Series

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
				Label: utils.Ternary(chapter.Title == "",
					m.Title()+" "+chapter.Label(),
					chapter.Label()),
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

func (m *manga) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return m.repository.ChapterImages(ctx, chapter.Id)
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
	ci.Web = m.seriesInfo.RefUrl()

	m.WriteGenreAndTags(chapter, ci)

	return ci
}

func (m *manga) WriteGenreAndTags(chapter Chapter, ci *comicinfo.ComicInfo) {
	tags := utils.FlatMapMany(chapter.Tags, m.seriesInfo.Tags)

	var genres, blackList models.Tags
	p, err := m.preferences.GetWithTags()
	if err != nil {
		m.Log.Error().Err(err).Msg("failed to get mapped genre tags, not setting any genres")
		if !m.hasWarnedBlacklist {
			m.hasWarnedBlacklist = true
			m.Notifier.NotifyContentQ(m.Title()+": Blacklist failed to load",
				fmt.Sprintf("Blacklist failed to load while writing ComicInfo, no tags or genres will be included."+
					" Check logs for full list of failed chapters"),
				models.Orange)
		}
	} else {
		genres = p.DynastyGenreTags
		blackList = p.BlackListedTags
	}

	tagContains := func(slice models.Tags, tag Tag) bool {
		return slice.Contains(tag.Id) || slice.Contains(tag.DisplayName)
	}

	tagAllowed := func(tag Tag) bool {
		return err == nil && !tagContains(blackList, tag)
	}

	ci.Genre = strings.Join(utils.MaybeMap(tags, func(t Tag) (string, bool) {
		if tagContains(genres, t) && tagAllowed(t) {
			return t.DisplayName, true
		}
		m.Log.Trace().Str("tag", t.DisplayName).
			Msg("ignoring tag as genre, not configured in preferences or blacklisted")
		return "", false
	}), ",")

	if m.Req.GetBool(IncludeNotMatchedTagsKey, false) {
		ci.Tags = strings.Join(utils.MaybeMap(tags, func(t Tag) (string, bool) {
			if !tagAllowed(t) {
				return "", false
			}
			if tagContains(genres, t) {
				return "", false
			}
			return t.DisplayName, true
		}), ",")
	} else {
		m.Log.Trace().Msg("not including unmatched tags in comicinfo.xml")
	}
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
