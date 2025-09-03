package mangadex

import (
	"context"
	"fmt"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/go-metroninfo"
	"github.com/rs/zerolog"
)

func (m *manga) WriteContentMetaData(ctx context.Context, chapter ChapterSearchData) error {
	metaPath := m.ContentPath(chapter)

	l := m.ContentLogger(chapter)

	err := m.fs.MkdirAll(metaPath, 0755)
	if err != nil {
		return err
	}

	if m.Req.GetBool(core.IncludeCover, true) {
		if err = m.writeCover(ctx, l, chapter); err != nil {
			return err
		}
	}

	l.Trace().Msg("writing comicinfoxml")
	if err = comicinfo.Save(m.fs, m.comicInfo(chapter), path.Join(metaPath, "comicinfo.xml")); err != nil {
		return err
	}

	/*l.Trace().Msg("writing MetronInfo.xml")
	if err = m.metronInfo(chapter).Save(path.Join(metaPath, "MetronInfo.xml"), true); err != nil {
		return err
	}*/

	return nil
}

func (m *manga) writeCover(ctx context.Context, l zerolog.Logger, chapter ChapterSearchData) error {
	toWrite, isFirstPage := m.getChapterCover(ctx, chapter)
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
	return m.fs.WriteFile(filePath, toWrite, 0644)
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
		Name:      m.SeriesInfo.Attributes.LangTitle(m.language),
		StartYear: m.SeriesInfo.Attributes.Year,
		AlternativeNames: utils.FlatMap(utils.Map(m.SeriesInfo.Attributes.AltTitles, func(t map[string]string) []metroninfo.AlternativeName {
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
	mi.Summary = m.markdownService.MdToSafeHtml(m.SeriesInfo.Attributes.LangDescription(m.language))
	mi.AgeRating = m.SeriesInfo.Attributes.ContentRating.MetronInfoAgeRating()
	mi.URLs = utils.Map(m.SeriesInfo.FormattedLinks(), func(t string) metroninfo.URL {
		return metroninfo.URL{
			Primary: t == m.SeriesInfo.RefUrl(),
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

	if m.SeriesInfo.Attributes.Status == StatusCompleted {
		switch {
		case m.lastFoundVolume == 0 && m.foundLastChapter:
			mi.Series.VolumeCount = int(math.Floor(m.lastFoundChapter))
		case m.foundLastChapter && m.foundLastVolume:
			mi.Series.VolumeCount = int(math.Floor(m.lastFoundVolume))
		case !m.hasWarned:
			m.hasWarned = true
			m.Log.Warn().
				Str("lastChapter", m.SeriesInfo.Attributes.LastChapter).
				Bool("foundLastChapter", m.foundLastChapter).
				Str("lastVolume", m.SeriesInfo.Attributes.LastVolume).
				Bool("foundLastVolume", m.foundLastVolume).
				Msg("Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?")
		}
	}

	mi.Genres = utils.MaybeMap(m.SeriesInfo.Attributes.Tags, func(t TagData) (metroninfo.Genre, bool) {
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
	mi.Tags = utils.MaybeMap(m.SeriesInfo.Attributes.Tags, func(t TagData) (metroninfo.Tag, bool) {
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

	authors := utils.Map(m.SeriesInfo.Authors(), roleMapper(metroninfo.RoleWriter))
	artists := utils.Map(m.SeriesInfo.Artists(), roleMapper(metroninfo.RoleArtist))
	scanlation := utils.Map(m.SeriesInfo.ScanlationGroup(), roleMapper(metroninfo.RoleTranslator))

	mi.Credits = utils.FlatMapMany(authors, artists, scanlation)
	mi.Notes = metronInfoNote
	now := time.Now()
	mi.LastModified = &now

	return mi
}

func (m *manga) comicInfo(chapter ChapterSearchData) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = utils.NonEmpty(m.Req.GetStringOrDefault(core.TitleOverride, ""), m.SeriesInfo.Attributes.LangTitle(m.language))
	ci.Year = m.SeriesInfo.Attributes.Year
	ci.Summary = m.markdownService.MdToSafeHtml(m.SeriesInfo.Attributes.LangDescription(m.language))
	ci.Manga = comicinfo.MangaYes
	ci.AgeRating = m.getAgeRating()
	ci.Web = strings.Join(m.SeriesInfo.FormattedLinks(), ",")
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

	alts := m.SeriesInfo.Attributes.LangAltTitles(m.language)
	if len(alts) > 0 {
		ci.LocalizedSeries = alts[0]
	}

	// OneShots do not have status
	if chapter.Attributes.Chapter != "" {
		if count, ok := m.getCiStatus(); ok {
			ci.Count = count
		}
	} else {
		ci.Format = "Special"
	}

	if v, err := strconv.Atoi(chapter.Attributes.Volume); err == nil {
		ci.Volume = v
	} else {
		m.Log.Trace().Err(err).Str("volume", chapter.Attributes.Volume).Msg("unable to parse volume number")
	}

	if chapter.Attributes.Chapter != "" {
		ci.Number = chapter.Attributes.Chapter
	}

	m.writeTagsAndGenres(ci)

	ci.Writer = strings.Join(m.SeriesInfo.Authors(), ",")
	ci.Colorist = strings.Join(m.SeriesInfo.Artists(), ",")

	ci.Notes = comicInfoNote
	return ci
}

func (m *manga) getAgeRating() comicinfo.AgeRating {
	tags := utils.MaybeMap(m.SeriesInfo.Attributes.Tags, func(t TagData) (core.Tag, bool) {
		tag, ok := t.Attributes.Name[m.language]
		if !ok {
			return nil, false
		}
		return core.NewStringTag(tag), true
	})

	mar := m.SeriesInfo.Attributes.ContentRating.ComicInfoAgeRating()
	ar, ok := m.GetAgeRating(tags)
	if !ok {
		return mar
	}

	highest := max(comicinfo.AgeRatingIndex[ar], comicinfo.AgeRatingIndex[mar])
	return comicinfo.IndexToAgeRating[highest]
}

// Mangadex has its own writeTagsAndGenres as they do have a concept of genre's and tag. As opposed to only tags

//nolint:funlen
func (m *manga) writeTagsAndGenres(ci *comicinfo.ComicInfo) {
	if m.Preference == nil {
		m.Log.Warn().Msg("No genres or tags will be set, blacklist couldn't be loaded")
		m.WarnPreferencesFailedToLoad()
		return
	}

	var blackList models.Tags = m.Preference.BlackListedTags

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

	ci.Genre = strings.Join(utils.MaybeMap(m.SeriesInfo.Attributes.Tags, func(t TagData) (string, bool) {
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

	ci.Tags = strings.Join(utils.MaybeMap(m.SeriesInfo.Attributes.Tags, func(t TagData) (string, bool) {
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
}

// getCiStatus updates the ComicInfo.Count field according the Mangadex's information
// and adds a notification in case a subscription has been exhausted
func (m *manga) getCiStatus() (int, bool) {
	if m.SeriesInfo.Attributes.Status != StatusCompleted {
		m.Log.Trace().Msg("Series not completed, no status to write")
		return 0, false
	}

	if m.SeriesInfo.Attributes.LastVolume == "" && m.SeriesInfo.Attributes.LastChapter == "" {
		m.Log.Warn().Msg("Mangadex marked this series as completed, but no last volume or chapter were provided?")
		return 0, false
	}

	var lastWantedChapter, lastWantedVolume float64
	if m.SeriesInfo.Attributes.LastChapter != "" {
		val, err := strconv.ParseFloat(m.SeriesInfo.Attributes.LastChapter, 64)
		if err != nil {
			m.Log.Warn().Err(err).Str("chapter", m.SeriesInfo.Attributes.LastChapter).
				Msg("Series was completed, but we failed to parse the last chapter from mangadex")
			return 0, false
		}
		lastWantedChapter = val
	}
	if m.SeriesInfo.Attributes.LastVolume != "" {
		val, err := strconv.ParseFloat(m.SeriesInfo.Attributes.LastVolume, 64)
		if err != nil {
			m.Log.Warn().Err(err).Str("volume", m.SeriesInfo.Attributes.LastVolume).
				Msg("Series was completed, but we failed to parse the last volume from mangadex")
			return 0, false
		}
		lastWantedVolume = val
	}

	var count, found float64

	if m.SeriesInfo.Attributes.LastVolume == "" && m.SeriesInfo.Attributes.LastChapter != "" {
		count = lastWantedChapter
		found = m.lastFoundChapter
	} else {
		count = lastWantedVolume
		found = m.lastFoundVolume
	}

	intCount := int(math.Floor(count))

	lastVolumeReachedChaptersMissing := m.SeriesInfo.Attributes.LastVolume != "" && m.SeriesInfo.Attributes.LastChapter != "" &&
		m.lastFoundChapter < lastWantedChapter
	if found < count || lastVolumeReachedChaptersMissing {
		if !m.hasWarned {
			m.hasWarned = true
			m.Log.Warn().
				Str("lastChapter", m.SeriesInfo.Attributes.LastChapter).
				Float64("lastFoundChapter", m.lastFoundChapter).
				Str("lastVolume", m.SeriesInfo.Attributes.LastVolume).
				Float64("lastFoundVolume", m.lastFoundVolume).
				Msg("Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?")
		}
		return intCount, true
	}

	total := fmt.Sprintf("%s Volumes, %s Chapters", m.SeriesInfo.Attributes.LastVolume, m.SeriesInfo.Attributes.LastChapter)
	m.NotifySubscriptionExhausted(total)
	return intCount, true
}

// getChapterCover returns the cover for the chapter, and if it's the first page in the chapter
// if no cover is found. Returns nil
func (m *manga) getChapterCover(ctx context.Context, chapter ChapterSearchData) ([]byte, bool) {
	l := m.ContentLogger(chapter)
	cover, ok := m.coverFactory(chapter.Attributes.Volume)
	if !ok {
		l.Debug().Msg("unable to find cover")
		return nil, false
	}

	coverBytes, isFirstPage, err := m.getBetterChapterCover(ctx, chapter, cover)
	if err != nil {
		l.Warn().Err(err).Msg("an error occurred when trying to compare cover with the first page. Falling back")
		coverBytes = cover.Bytes
	}
	return coverBytes, isFirstPage
}

// getBetterChapterCover check if a higher quality cover is used inside chapters. Returns true
// when the Cover returned if the first page of the chapter passed as an argument
func (m *manga) getBetterChapterCover(ctx context.Context, chapter ChapterSearchData, currentCover *Cover) ([]byte, bool, error) {
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

	res, err := m.repository.GetChapterImages(ctx, chapter.Id)
	if err != nil {
		return nil, false, err
	}

	images := res.FullImageUrls()

	if len(images) == 0 {
		return currentCover.Bytes, false, nil
	}

	candidateBytes, err := m.Download(ctx, images[0])
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
