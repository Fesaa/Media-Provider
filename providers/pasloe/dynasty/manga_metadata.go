package dynasty

import (
	"context"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"path"
	"slices"
	"strconv"
	"strings"
)

func (m *manga) WriteContentMetaData(chapter Chapter) error {

	if m.Req.GetBool(IncludeCover, true) {
		if err := m.writeCover(chapter); err != nil {
			return err
		}
	}

	m.Log.Trace().Str("chapter", chapter.Chapter).Msg("writing comicinfoxml")
	return comicinfo.Save(m.comicInfo(chapter), path.Join(m.ContentPath(chapter), "ComicInfo.xml"))
}

func (m *manga) writeCover(chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover.jpg")

	if !m.hasCheckedCover {
		m.hasCheckedCover = true
		if err := m.tryReplaceCover(); err != nil {
			return err
		}
	}

	if len(m.coverBytes) == 0 {
		m.Log.Trace().Str("chapter", chapter.Chapter).Msg("no cover bytes set, downloading from url")
		return m.downloadAndWrite(m.seriesInfo.CoverUrl, filePath)
	}

	return m.fs.WriteFile(filePath, m.coverBytes, 0755)
}

func (m *manga) tryReplaceCover() error {
	m.Log.Trace().Msg("Checking if first image of first chapter has a higher quality cover")
	firstChapter := utils.Find(m.seriesInfo.Chapters, func(chapter Chapter) bool {
		return chapter.Chapter == "1"
	})

	if firstChapter == nil {
		return nil
	}

	// TODO: Pass context
	images, err := m.repository.ChapterImages(context.Background(), firstChapter.Id)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return nil
	}

	coverBytes, err := m.download(m.seriesInfo.CoverUrl)
	if err != nil {
		return err
	}

	firstChapterCoverBytes, err := m.download(images[0])
	if err != nil {
		return err
	}

	// Dynasty doesn't have a concept for per chapter/volume covers. So we're using one cover at all times anyway
	// Should set IncludeCover metadata to false if you want to use chapter covers and add your own later
	m.coverBytes, _, err = m.imageService.Better(coverBytes, firstChapterCoverBytes)
	if err != nil {
		return err
	}

	return nil
}

func (m *manga) comicInfo(chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = m.seriesInfo.Title
	ci.AlternateSeries = m.seriesInfo.AltTitle
	ci.Summary = m.markdownService.SanitizeHtml(m.seriesInfo.Description)
	ci.Manga = comicinfo.MangaYes
	ci.Title = chapter.Title
	if chapter.Volume != "" {
		if vol, err := strconv.Atoi(chapter.Volume); err == nil {
			ci.Volume = vol
		} else {
			m.Log.Trace().Err(err).Str("chapter", chapter.Volume).Msg("could not convert volume to int")
		}
	}

	ci.Writer = strings.Join(utils.Map(m.seriesInfo.Authors, func(t Author) string {
		return t.DisplayName
	}), ",")
	ci.Web = m.seriesInfo.RefUrl()

	m.WriteGenreAndTags(chapter, ci)

	if ar, ok := m.getAgeRating(chapter); ok {
		ci.AgeRating = ar
	}

	return ci
}

func (m *manga) WriteGenreAndTags(chapter Chapter, ci *comicinfo.ComicInfo) {
	tags := utils.FlatMapMany(chapter.Tags, m.seriesInfo.Tags)

	var genres, blackList models.Tags
	p, err := m.preferences.GetComplete()
	if err != nil {
		m.Log.Error().Err(err).Msg("failed to get mapped genre tags, not setting any genres")
		if !m.hasWarnedBlacklist {
			m.hasWarnedBlacklist = true
			m.Notifier.NotifyContentQ(
				m.TransLoco.GetTranslation("blacklist-failed-to-load-title", m.Title()),
				m.TransLoco.GetTranslation("blacklist-failed-to-load-summary"),
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

func (m *manga) getAgeRating(chapter Chapter) (comicinfo.AgeRating, bool) {
	if m.Preference == nil {
		m.Log.Warn().Msg("Could not load age rate mapping, not setting age rating")
		return "", false
	}

	var mappings models.AgeRatingMappings = m.Preference.AgeRatingMappings
	allTags := append(m.seriesInfo.Tags, chapter.Tags...) //nolint:gocritic
	weights := utils.MaybeMap(allTags, func(t Tag) (int, bool) {
		ar, ok := mappings.GetAgeRating(t.DisplayName)
		if !ok {
			return 0, false
		}

		return comicinfo.AgeRatingIndex[ar], true
	})

	if len(weights) == 0 {
		return "", false
	}

	return comicinfo.IndexToAgeRating[slices.Max(weights)], true
}
