package mangadex

import (
	"context"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/go-metroninfo"
	"github.com/rs/zerolog"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

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
	toWrite, isFirstPage := m.getChapterCover(chapter)
	if toWrite == nil {
		l.Trace().Msg("no cover found")
		return nil
	}
	if isFirstPage {
		l.Trace().Msg("first page is the cover, not writing cover again")
		return nil
	}
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover.jpg")
	return os.WriteFile(filePath, toWrite, 0644)
}

// metronInfo DO NOT USE: Code is outdated!
//
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
	if m.Preference == nil {
		m.Log.Warn().Msg("No genres or tags will be set, blacklist couldn't be loaded")
		if !m.hasWarnedBlacklist {
			m.hasWarnedBlacklist = true
			m.Notifier.NotifyContentQ(
				m.TransLoco.GetTranslation("blacklist-failed-to-load-title", m.Title()),
				m.TransLoco.GetTranslation("blacklist-failed-to-load-summary"),
				models.Orange)
		}
	} else {
		blackList = m.Preference.BlackListedTags
	}

	tagAllowed := func(tag TagData, name string) bool {
		if m.Preference == nil {
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

// writeCIStatus updates the ComicInfo.Count field according the Mangadex's information
// and adds a notification in case a subscription has been exhausted
func (m *manga) writeCIStatus(ci *comicinfo.ComicInfo) {
	if m.info.Attributes.Status != StatusCompleted {
		m.Log.Trace().Msg("Series not completed, no status to write")
		return
	}

	if m.info.Attributes.LastVolume == "" && m.info.Attributes.LastChapter == "" {
		m.Log.Warn().Msg("Mangadex marked this series as completed, but no last volume or chapter were provided?")
		return
	}

	var count, found int
	var content string
	if m.info.Attributes.LastVolume == "" && m.info.Attributes.LastChapter != "" {
		val, err := strconv.ParseInt(m.info.Attributes.LastChapter, 10, 64)
		if err != nil {
			m.Log.Warn().Err(err).Str("chapter", m.info.Attributes.LastChapter).
				Msg("Series was completed, but we failed to parse the last chapter from mangadex")
			return
		}
		count = int(val)
		found = m.lastFoundChapter
		content = "Chapters"
	} else {
		val, err := strconv.ParseInt(m.info.Attributes.LastVolume, 10, 64)
		if err != nil {
			m.Log.Warn().Err(err).Str("volume", m.info.Attributes.LastVolume).
				Msg("Series was completed, but we failed to parse the last volume from mangadex")
			return
		}
		count = int(val)
		found = m.lastFoundVolume
		content = "Volumes"
	}

	ci.Count = count
	if found < count {
		if !m.hasWarned {
			m.hasWarned = true
			m.Log.Warn().
				Str("lastChapter", m.info.Attributes.LastChapter).
				Bool("foundLastChapter", m.foundLastChapter).
				Str("lastVolume", m.info.Attributes.LastVolume).
				Bool("foundLastVolume", m.foundLastVolume).
				Msg("Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?")
		}
		return
	}

	// Series has completed, and everything has been downloaded
	if !m.Req.IsSubscription || m.hasNotifiedSub {
		return
	}

	m.hasNotifiedSub = true
	m.Notifier.NotifyContent(m.TransLoco.GetTranslation("sub-downloaded-all-title"),
		m.Title(), m.TransLoco.GetTranslation("sub-downloaded-all", m.Title(), count, content))
}

// getChapterCover returns the cover for the chapter, and if it's the first page in the chapter
// if no cover is found. Returns nil
func (m *manga) getChapterCover(chapter ChapterSearchData) ([]byte, bool) {
	l := m.ContentLogger(chapter)
	cover, ok := m.coverFactory(chapter.Attributes.Volume)
	if !ok {
		l.Debug().Msg("unable to find cover")
		return nil, false
	}

	coverBytes, isFirstPage, err := m.getBetterChapterCover(chapter, cover)
	if err != nil {
		l.Warn().Err(err).Msg("an error occurred when trying to compare cover with the first page. Falling back")
		coverBytes = cover.Bytes
	}
	return coverBytes, isFirstPage
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
