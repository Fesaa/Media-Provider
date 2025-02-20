package mangadex

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
	"github.com/Fesaa/go-metroninfo"
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

const comicInfoNote = "This comicinfo.xml was auto generated by Media-Provider, with information from mangadex. Source code can be found here: https://github.com/Fesaa/Media-Provider/"
const metronInfoNote = "This metroninfo.xml was auto generated by Media-Provider, with information from mangadex. Source code can be found here: https://github.com/Fesaa/Media-Provider/"
const timeLayout = "2006-01-02T15:04:05Z07:00"
const genreTag = "genre"

func NewManga(scope *dig.Scope) api.Downloadable {
	var m *manga

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, httpClient *http.Client,
		repository Repository, markdownService services.MarkdownService,
		preferences models.Preferences, imageService services.ImageService,
	) {
		m = &manga{
			id:              req.Id,
			httpClient:      httpClient,
			repository:      repository,
			markdownService: markdownService,
			volumeMetadata:  make([]string, 0),
			preferences:     preferences,
			imageService:    imageService,

			language: utils.MustHave(req.GetString(LanguageKey, "en")),
		}
		m.DownloadBase = api.NewDownloadableFromBlock[ChapterSearchData](scope, "mangadex", m)
	}))

	return m
}

type manga struct {
	*api.DownloadBase[ChapterSearchData]
	id string

	httpClient      *http.Client
	repository      Repository
	markdownService services.MarkdownService
	preferences     models.Preferences
	imageService    services.ImageService

	info     *MangaSearchData
	chapters ChapterSearchResponse

	coverFactory   CoverFactory
	volumeMetadata []string

	lastFoundChapter int
	lastFoundVolume  int
	foundLastVolume  bool
	foundLastChapter bool

	hasWarned          bool
	hasWarnedBlacklist bool

	language string
}

func (m *manga) Title() string {
	if m.info == nil {
		if m.TempTitle != "" {
			return m.TempTitle
		}

		return m.id
	}

	return m.info.Attributes.LangTitle(m.language)
}

func (m *manga) Provider() models.Provider {
	return m.Req.Provider
}

func (m *manga) RefUrl() string {
	if m.info == nil {
		return ""
	}

	return m.info.RefURL()
}

func (m *manga) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		mangaInfo, err := m.repository.GetManga(ctx, m.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading manga info")
			}
			m.Cancel()
			close(out)
			return
		}
		m.info = &mangaInfo.Data

		chapters, err := m.repository.GetChapters(ctx, m.id)
		if err != nil || chapters == nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading chapter info")
			}
			m.Cancel()
			close(out)
			return
		}

		m.chapters = m.FilterChapters(chapters)
		m.SetSeriesStatus()

		if m.Req.GetBool(IncludeCover, true) {
			covers, err := m.repository.GetCoverImages(ctx, m.id)
			if err != nil || covers == nil {
				m.Log.Warn().Err(err).Msg("error while loading manga coverFactory, ignoring")
				m.coverFactory = defaultCoverFactory
			} else {
				m.coverFactory = m.getCoverFactoryLang(covers)
			}
		}

		close(out)
	}()
	return out
}

func (m *manga) SetSeriesStatus() {
	var maxVolume int64 = -1
	var maxChapter int64 = -1

	// If there is a last chapter present, but no last volume is given. We assume that the series does not use volumes
	m.foundLastVolume = m.info.Attributes.LastVolume == "" && m.info.Attributes.LastChapter != ""
	for _, ch := range m.chapters.Data {
		if ch.Attributes.Volume == m.info.Attributes.LastVolume && m.info.Attributes.LastVolume != "" {
			m.foundLastVolume = true
		}
		if ch.Attributes.Chapter == m.info.Attributes.LastChapter && m.info.Attributes.LastChapter != "" {
			m.foundLastChapter = true
		}

		if val, err := strconv.ParseInt(ch.Attributes.Volume, 10, 64); err == nil {
			maxVolume = max(maxVolume, val)
		} else {
			m.Log.Trace().Str("volume", ch.Attributes.Volume).Str("chapter", ch.Attributes.Chapter).
				Msg("not adding chapter, as Volume string isn't an int")
		}

		if val, err := strconv.ParseInt(ch.Attributes.Chapter, 10, 64); err == nil {
			maxChapter = max(maxChapter, val)
		} else {
			m.Log.Trace().Str("volume", ch.Attributes.Volume).Str("chapter", ch.Attributes.Chapter).
				Msg("not adding chapter, as Chapter string isn't an int")
		}
	}
	// We can set these safely as they're only written when found
	m.lastFoundVolume = int(maxVolume)
	m.lastFoundChapter = int(maxChapter)
}

func (m *manga) FilterChapters(c *ChapterSearchResponse) ChapterSearchResponse {
	scanlation := func() string {
		if scanlationGroup, ok := m.Req.GetString(ScanlationGroupKey); ok {
			m.Log.Debug().Str("scanlationGroup", scanlationGroup).
				Msg("loading manga info, prioritizing chapters from a specific Scanlation group or user")
			return scanlationGroup
		}

		return ""
	}()
	chaptersMap := utils.GroupBy(c.Data, func(v ChapterSearchData) string {
		return v.Attributes.Chapter
	})

	newData := make([]ChapterSearchData, 0)
	for _, chapters := range chaptersMap {
		chapter := utils.Find(chapters, m.chapterSearchFunc(scanlation, true))

		// Retry by skipping scanlation check
		if chapter == nil && scanlation != "" {
			chapter = utils.Find(chapters, m.chapterSearchFunc("", true))
		}

		if chapter != nil {
			newData = append(newData, *chapter)
		}
	}

	if m.Req.GetBool(DownloadOneShotKey) {
		// OneShots do not have a chapter, so will be mapped under the empty string
		if chapters, ok := chaptersMap[""]; ok {
			newData = append(newData, utils.Filter(chapters, m.chapterSearchFunc(scanlation, false))...)
		}
	}

	c.Data = newData
	return *c
}

func (m *manga) chapterSearchFunc(scanlation string, skipOneShot bool) func(ChapterSearchData) bool {
	return func(data ChapterSearchData) bool {
		if data.Attributes.TranslatedLanguage != m.language {
			return false
		}
		// Skip over official publisher chapters, we cannot download these from mangadex
		if data.Attributes.ExternalUrl != "" {
			return false
		}

		if data.Attributes.Chapter == "" && skipOneShot {
			return false
		}

		if scanlation == "" {
			return true
		}

		return slices.ContainsFunc(data.Relationships, func(relationship Relationship) bool {
			if relationship.Type != "scanlation_group" && relationship.Type != "user" {
				return false
			}

			return relationship.Id == scanlation
		})
	}
}

func (m *manga) All() []ChapterSearchData {
	return m.chapters.Data
}

func (m *manga) ContentList() []payload.ListContentData {
	if len(m.chapters.Data) == 0 {
		return nil
	}

	data := utils.GroupBy(m.chapters.Data, func(v ChapterSearchData) string {
		return v.Attributes.Volume
	})

	childrenFunc := func(chapters []ChapterSearchData) []payload.ListContentData {
		slices.SortFunc(chapters, func(a, b ChapterSearchData) int {
			if a.Attributes.Volume != b.Attributes.Volume {
				return (int)(b.Volume() - a.Volume())
			}
			return (int)(b.Chapter() - a.Chapter())
		})

		return utils.Map(chapters, func(chapter ChapterSearchData) payload.ListContentData {
			return payload.ListContentData{
				SubContentId: chapter.Id,
				Selected:     len(m.ToDownloadUserSelected) == 0 || slices.Contains(m.ToDownloadUserSelected, chapter.Id),
				Label: utils.Ternary(chapter.Attributes.Title == "",
					m.Title()+" "+chapter.Label(),
					chapter.Label()),
			}
		})
	}

	sortSlice := utils.Keys(data)
	slices.SortFunc(sortSlice, utils.SortFloats)

	out := make([]payload.ListContentData, 0, len(data))
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

func (m *manga) ContentDir(chapter ChapterSearchData) string {
	if chapter.Attributes.Chapter == "" {
		return fmt.Sprintf("%s OneShot %s", m.Title(), chapter.Attributes.Title)
	}

	if chpt, err := strconv.ParseFloat(chapter.Attributes.Chapter, 32); err == nil {
		chDir := fmt.Sprintf("%s Ch. %s", m.Title(), utils.PadFloat(chpt, 4))
		return chDir
	} else if chapter.Attributes.Chapter != "" { // Don't warm for empty chpt. They're expected to fail
		m.Log.Warn().Err(err).Str("chapter", chapter.Attributes.Chapter).Msg("unable to parse chpt number, not padding")
	}

	return fmt.Sprintf("%s Ch. %s", m.Title(), chapter.Attributes.Chapter)
}

func (m *manga) ContentPath(chapter ChapterSearchData) string {
	base := path.Join(m.Client.GetBaseDir(), m.GetBaseDir(), m.Title())
	if chapter.Attributes.Volume == "" {
		return path.Join(base, m.ContentDir(chapter))
	}
	return path.Join(base, m.volumeDir(chapter.Attributes.Volume), m.ContentDir(chapter))
}

func (m *manga) ContentKey(chapter ChapterSearchData) string {
	return chapter.Id
}

func (m *manga) ContentLogger(chapter ChapterSearchData) zerolog.Logger {
	builder := m.Log.With().
		Str("chapterId", chapter.Id).
		Str("chapter", chapter.Attributes.Chapter)

	if chapter.Attributes.Volume != "" {
		builder = builder.Str("volume", chapter.Attributes.Volume)
	}

	if chapter.Attributes.Title != "" {
		builder = builder.Str("title", chapter.Attributes.Title)
	}

	return builder.Logger()
}

func (m *manga) ContentUrls(ctx context.Context, chapter ChapterSearchData) ([]string, error) {
	imageInfo, err := m.repository.GetChapterImages(ctx, chapter.Id)
	if err != nil {
		return nil, err
	}
	return imageInfo.FullImageUrls(), nil
}

func (m *manga) WriteContentMetaData(chapter ChapterSearchData) error {
	metaKey, metaPath := chapter.Attributes.Chapter, m.ContentPath(chapter)

	if slices.Contains(m.volumeMetadata, metaKey) {
		m.Log.Trace().
			Str("volume", chapter.Attributes.Volume).
			Str("chapter", chapter.Attributes.Chapter).
			Msg("volume metadata already written, skipping")
		return nil
	}

	l := m.ContentLogger(chapter)

	err := os.MkdirAll(metaPath, 0755)
	if err != nil {
		return err
	}

	if m.Req.GetBool(IncludeCover, true) {
		if err = m.writeCover(l, chapter); err != nil {
			return err
		}
	}

	l.Trace().Msg("writing comicinfoxml")
	if err = comicinfo.Save(m.comicInfo(chapter), path.Join(metaPath, "comicinfo.xml")); err != nil {
		return err
	}

	/*l.Trace().Msg("writing MetronInfo.xml")
	if err = m.metronInfo(chapter).Save(path.Join(metaPath, "MetronInfo.xml"), true); err != nil {
		return err
	}*/

	m.volumeMetadata = append(m.volumeMetadata, metaKey)
	return nil
}

func (m *manga) writeCover(l zerolog.Logger, chapter ChapterSearchData) error {
	cover, ok := m.coverFactory(chapter.Attributes.Volume)
	if !ok {
		l.Debug().Msg("unable to find cover")
		return nil
	}

	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover.jpg")
	toWrite, isFirstPage, err := m.getBetterChapterCover(chapter, cover)
	if err != nil {
		l.Warn().Err(err).Msg("an error occurred when trying to compare cover with the first page. Falling back")
		toWrite = cover.Bytes
	}

	if isFirstPage && err == nil {
		l.Trace().Msg("first page is the cover, not writing cover again")
		return nil
	}

	return os.WriteFile(filePath, toWrite, 0644)
}

// getBetterChapterCover check if a higher quality cover is used inside chapters. Returns true
// when the Cover returned if the first page of the chapter passed as an argument
func (m *manga) getBetterChapterCover(chapter ChapterSearchData, currentCover *Cover) ([]byte, bool, error) {
	replaced := false
	if chapter.Attributes.Volume != "" {
		chapters := utils.GroupBy(m.chapters.Data, func(v ChapterSearchData) string {
			return v.Attributes.Volume
		})[chapter.Attributes.Volume]

		slices.SortFunc(chapters, func(a, b ChapterSearchData) int {
			return (int)(a.Volume() - b.Volume())
		})

		if chapter.Id != chapters[0].Id {
			m.Log.Trace().
				Str("originalChapter", chapter.Attributes.Chapter).
				Str("newChapter", chapters[0].Attributes.Chapter).
				Msg("overwriting chapter to check cover of")
			chapter = chapters[0]
			replaced = true
		}
	}

	res, err := m.repository.GetChapterImages(context.Background(), chapter.Id)
	if err != nil {
		return nil, false, err
	}

	images := res.FullImageUrls()

	if len(images) == 0 {
		return currentCover.Bytes, false, nil
	}

	candidateBytes, err := m.download(images[0])
	if err != nil {
		return nil, false, err
	}

	better, replacedBytes, err := m.imageService.Better(currentCover.Bytes, candidateBytes)
	if err != nil {
		return nil, false, err
	}

	// If the chapter was replaced, should still write the cover
	return better, !replaced && replacedBytes, nil
}

//nolint:funlen
func (m *manga) metronInfo(chapter ChapterSearchData) *metroninfo.MetronInfo {
	mi := metroninfo.NewMetronInfo()

	mi.IDS = []metroninfo.ID{
		{
			Source:  metroninfo.SourceMangaDex,
			Primary: true,
			Value:   m.id,
		},
	}

	mi.Series = metroninfo.Series{
		Name:      m.info.Attributes.LangTitle(m.language),
		StartYear: m.info.Attributes.Year,
		AlternativeNames: utils.FlatMap(utils.Map(m.info.Attributes.AltTitles, func(t map[string]string) []metroninfo.AlternativeName {
			var out []metroninfo.AlternativeName
			for key, value := range t {
				out = append(out, metroninfo.AlternativeName{
					Lang:  metroninfo.LanguageCode(key),
					Value: value,
				})
			}
			return out
		})),
	}
	mi.Summary = m.markdownService.MdToSafeHtml(m.info.Attributes.LangDescription(m.language))
	mi.AgeRating = m.info.Attributes.ContentRating.MetronInfoAgeRating()
	mi.URLs = utils.Map(m.info.FormattedLinks(), func(t string) metroninfo.URL {
		return metroninfo.URL{
			Primary: t == m.info.RefURL(),
			Value:   t,
		}
	})

	if m.lastFoundVolume == 0 {
		mi.Stories = metroninfo.Stories{{Value: chapter.Attributes.Title}}

		if chapter.Attributes.PublishedAt != "" {
			publishTime, err := time.Parse(timeLayout, chapter.Attributes.PublishedAt)
			if err != nil {
				m.Log.Warn().Err(err).
					Str("chapter", chapter.Attributes.Chapter).
					Msg("unable to parse published date")
			} else {
				mi.StoreDate = (*metroninfo.Date)(&publishTime)
			}
		}
	}

	if m.info.Attributes.Status == StatusCompleted {
		switch {
		case m.lastFoundVolume == 0 && m.foundLastChapter:
			mi.Series.VolumeCount = m.lastFoundChapter
		case m.foundLastChapter && m.foundLastVolume:
			mi.Series.VolumeCount = m.lastFoundVolume
		case !m.hasWarned:
			m.hasWarned = true
			m.Log.Warn().
				Str("lastChapter", m.info.Attributes.LastChapter).
				Bool("foundLastChapter", m.foundLastChapter).
				Str("lastVolume", m.info.Attributes.LastVolume).
				Bool("foundLastVolume", m.foundLastVolume).
				Msg("Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?")
		}
	}

	mi.Genres = utils.MaybeMap(m.info.Attributes.Tags, func(t TagData) (metroninfo.Genre, bool) {
		n, ok := t.Attributes.Name[m.language]
		if !ok {
			return metroninfo.Genre{}, false
		}

		if t.Attributes.Group != genreTag {
			return metroninfo.Genre{}, false
		}

		return metroninfo.Genre{
			Value: n,
		}, true
	})
	mi.Tags = utils.MaybeMap(m.info.Attributes.Tags, func(t TagData) (metroninfo.Tag, bool) {
		n, ok := t.Attributes.Name[m.language]
		if !ok {
			return metroninfo.Tag{}, false
		}

		if t.Attributes.Group == genreTag {
			return metroninfo.Tag{}, false
		}

		return metroninfo.Tag{
			Value: n,
		}, true
	})

	roleMapper := func(r metroninfo.RoleValue) func(t string) metroninfo.Credit {
		return func(t string) metroninfo.Credit {
			return metroninfo.Credit{
				Creator: metroninfo.Resource{
					Value: t,
				},
				Roles: []metroninfo.Role{{
					Value: r,
				}},
			}
		}
	}

	authors := utils.Map(m.info.Authors(), roleMapper(metroninfo.RoleWriter))
	artists := utils.Map(m.info.Artists(), roleMapper(metroninfo.RoleArtist))
	scanlation := utils.Map(m.info.ScanlationGroup(), roleMapper(metroninfo.RoleTranslator))

	mi.Credits = utils.FlatMapMany(authors, artists, scanlation)
	mi.Notes = metronInfoNote
	now := time.Now()
	mi.LastModified = &now

	return mi
}

//nolint:funlen
func (m *manga) comicInfo(chapter ChapterSearchData) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = m.info.Attributes.LangTitle(m.language)
	ci.Year = m.info.Attributes.Year
	ci.Summary = m.markdownService.MdToSafeHtml(m.info.Attributes.LangDescription(m.language))
	ci.Manga = comicinfo.MangaYes
	ci.AgeRating = m.info.Attributes.ContentRating.ComicInfoAgeRating()
	ci.Web = strings.Join(m.info.FormattedLinks(), ",")
	ci.LanguageISO = m.language

	ci.Title = chapter.Attributes.Title
	if chapter.Attributes.PublishedAt != "" {
		publishTime, err := time.Parse(timeLayout, chapter.Attributes.PublishedAt)
		if err != nil {
			m.Log.Warn().Err(err).Str("chapter", chapter.Attributes.Chapter).Msg("unable to parse published date")
		} else {
			ci.Year = publishTime.Year()
			ci.Month = int(publishTime.Month())
			ci.Day = publishTime.Day()
		}
	}

	alts := m.info.Attributes.LangAltTitles(m.language)
	if len(alts) > 0 {
		ci.LocalizedSeries = alts[0]
	}

	// OneShots do not have status
	if chapter.Attributes.Chapter != "" {
		m.writeCIStatus(ci)
	}

	if v, err := strconv.Atoi(chapter.Attributes.Volume); err == nil {
		ci.Volume = v
	} else {
		m.Log.Trace().Err(err).Str("volume", chapter.Attributes.Volume).Msg("unable to parse volume number")
	}

	var blackList models.Tags
	p, err := m.preferences.GetWithTags()
	if err != nil {
		m.Log.Error().Err(err).Msg("No genres or tags will be set, blacklist couldn't be loaded")

		if !m.hasWarnedBlacklist {
			m.hasWarnedBlacklist = true
			m.Notifier.NotifyContentQ(
				m.TransLoco.GetTranslation("blacklist-failed-to-load-title", m.Title()),
				m.TransLoco.GetTranslation("blacklist-failed-to-load-summary"),
				models.Orange)
		}
	} else {
		blackList = p.BlackListedTags
	}

	tagAllowed := func(tag TagData, name string) bool {
		if err != nil {
			return false
		}

		if blackList.Contains(name) {
			return false
		}

		if blackList.Contains(tag.Id) {
			return false
		}
		return true
	}

	ci.Genre = strings.Join(utils.MaybeMap(m.info.Attributes.Tags, func(t TagData) (string, bool) {
		n, ok := t.Attributes.Name[m.language]
		if !ok {
			return "", false
		}

		if t.Attributes.Group != genreTag {
			return "", false
		}

		if !tagAllowed(t, n) {
			return "", false
		}

		return n, true
	}), ",")

	ci.Tags = strings.Join(utils.MaybeMap(m.info.Attributes.Tags, func(t TagData) (string, bool) {
		n, ok := t.Attributes.Name[m.language]
		if !ok {
			return "", false
		}

		if t.Attributes.Group == genreTag {
			return "", false
		}

		if !tagAllowed(t, n) {
			return "", false
		}

		return n, true
	}), ",")

	ci.Writer = strings.Join(m.info.Authors(), ",")
	ci.Colorist = strings.Join(m.info.Artists(), ",")

	ci.Notes = comicInfoNote
	return ci
}

// writeStatus Add the comicinfo#count field if the manga has completed, so Kavita can add the correct Completed marker
// We can't add it for others, as mangadex is community sourced, so may lag behind. But this should be correct
func (m *manga) writeCIStatus(ci *comicinfo.ComicInfo) {
	if m.info.Attributes.Status != StatusCompleted {
		return
	}
	switch {
	case m.lastFoundVolume == 0 && m.foundLastChapter:
		ci.Count = m.lastFoundChapter
	case m.foundLastChapter && m.foundLastVolume:
		ci.Count = m.lastFoundVolume
	case !m.hasWarned:
		m.hasWarned = true
		m.Log.Warn().
			Str("lastChapter", m.info.Attributes.LastChapter).
			Bool("foundLastChapter", m.foundLastChapter).
			Str("lastVolume", m.info.Attributes.LastVolume).
			Bool("foundLastVolume", m.foundLastVolume).
			Msg("Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?")
	}
}

func (m *manga) DownloadContent(page int, chapter ChapterSearchData, url string) error {
	filePath := path.Join(m.ContentPath(chapter), fmt.Sprintf("page %s.jpg", utils.PadInt(page, 4)))
	if err := m.downloadAndWrite(url, filePath); err != nil {
		return err
	}
	m.ImagesDownloaded++
	return nil
}

var (
	contentRegex = regexp.MustCompile(".* (?:Ch|Vol)\\. ([\\d|\\.]+).cbz")
	oneShotRegex = regexp.MustCompile(".+ OneShot .+\\.cbz")
)

func (m *manga) IsContent(name string) bool {
	if contentRegex.MatchString(name) {
		return true
	}

	if oneShotRegex.MatchString(name) {
		return true
	}

	return false
}

func (m *manga) ShouldDownload(chapter ChapterSearchData) bool {
	// Backwards compatibility check if volume has been downloaded
	if _, ok := m.GetContentByName(m.volumeDir(chapter.Attributes.Volume) + ".cbz"); ok {
		return false
	}

	content, ok := m.GetContentByName(m.ContentDir(chapter) + ".cbz")
	if !ok {
		return true
	}

	// No extra I/O needing, empty volumes will never be replaced
	if chapter.Attributes.Volume == "" {
		return false
	}

	return m.replaceAndShouldDownload(chapter, content)
}

func (m *manga) replaceAndShouldDownload(chapter ChapterSearchData, content api.Content) bool {
	l := m.ContentLogger(chapter)
	fullPath := path.Join(m.Client.GetBaseDir(), content.Path)

	ci, err := comicinfo.ReadInZip(fullPath)
	if err != nil {
		l.Warn().Err(err).Str("path", fullPath).Msg("unable to read comic info in zip")
		return false
	}

	if strconv.Itoa(ci.Volume) == chapter.Attributes.Volume {
		l.Trace().Str("path", fullPath).Msg("Volume on disk matches, not replacing")
		return false
	}

	l.Debug().Int("onDiskVolume", ci.Volume).Str("path", fullPath).
		Msg("Loose chapter has been assigned to a volume, replacing")

	// Opted to remove, and redownload the entire chapter if the volume marker changes
	// One could argue that only the comicinfo.xml should be replaced if this happens.
	// Making the assumption that new content may be added in a chapter once it's added to a volume.
	// Especially the first, and last chapter of the volume.
	if err = os.Remove(fullPath); err != nil {
		l.Error().Err(err).Str("path", fullPath).Msg("unable to remove old chapter, not downloading new")
		return false
	}

	return true
}

func (m *manga) volumeDir(v string) string {
	if v == "" {
		return fmt.Sprintf("%s Special", m.Title())
	}

	return fmt.Sprintf("%s Vol. %s", m.Title(), v)
}

func (m *manga) download(url string, tryAgain ...bool) ([]byte, error) {
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			m.Log.Warn().Err(err).Msg("error closing body")
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	if len(tryAgain) > 0 && !tryAgain[0] {
		m.Log.Error().Msg("Reached rate limit, after sleeping. What is going on?")
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	retryAfter := resp.Header.Get("X-RateLimit-Retry-After")

	var d time.Duration
	if unix, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
		t := time.Unix(unix, 0)
		d = time.Until(t)
	} else {
		d = time.Minute
	}

	m.Log.Warn().Str("retryAfter", retryAfter).Dur("sleeping_for", d).Msg("Hit rate limit, try again after it's over")

	time.Sleep(d)
	return m.download(url, false)
}

func (m *manga) downloadAndWrite(url string, path string, tryAgain ...bool) error {
	data, err := m.download(url, tryAgain...)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, data, 0755); err != nil {
		return err
	}

	return nil
}
