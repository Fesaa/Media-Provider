package bato

import (
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
	return comicinfo.Save(m.fs, m.comicInfo(chapter), path.Join(m.ContentPath(chapter), "ComicInfo.xml"))
}

func (m *manga) writeCover(chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover.jpg")
	return m.downloadAndWrite(m.seriesInfo.CoverUrl, filePath)
}

func (m *manga) comicInfo(chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = m.seriesInfo.Title
	ci.AlternateSeries = m.seriesInfo.OriginalTitle
	ci.Summary = m.seriesInfo.Summary
	ci.Manga = comicinfo.MangaYes
	ci.Title = chapter.Title

	if chapter.Volume != "" {
		if vol, err := strconv.Atoi(chapter.Volume); err == nil {
			ci.Volume = vol
		} else {
			m.Log.Trace().Err(err).Str("chapter", chapter.Volume).Msg("could not convert volume to int")
		}
	}

	if chapter.Chapter != "" {
		ci.Number = chapter.Chapter
	}

	ci.Writer = strings.Join(m.seriesInfo.Authors, ",")
	ci.Web = m.seriesInfo.RefUrl()

	m.WriteGenreAndTags(ci)

	if ar, ok := m.getAgeRating(); ok {
		ci.AgeRating = ar
	}

	return ci

}

func (m *manga) WriteGenreAndTags(ci *comicinfo.ComicInfo) {
	tags := m.seriesInfo.Tags

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

	tagContains := func(slice models.Tags, tag string) bool {
		return slice.Contains(tag)
	}

	tagAllowed := func(tag string) bool {
		return err == nil && !tagContains(blackList, tag)
	}

	ci.Genre = strings.Join(utils.MaybeMap(tags, func(t string) (string, bool) {
		if tagContains(genres, t) && tagAllowed(t) {
			return t, true
		}
		m.Log.Trace().Str("tag", t).
			Msg("ignoring tag as genre, not configured in preferences or blacklisted")
		return "", false
	}), ",")

	if m.Req.GetBool(IncludeNotMatchedTagsKey, false) {
		ci.Tags = strings.Join(utils.MaybeMap(tags, func(t string) (string, bool) {
			if !tagAllowed(t) {
				return "", false
			}
			if tagContains(genres, t) {
				return "", false
			}
			return t, true
		}), ",")
	} else {
		m.Log.Trace().Msg("not including unmatched tags in comicinfo.xml")
	}
}

func (m *manga) getAgeRating() (comicinfo.AgeRating, bool) {
	if m.Preference == nil {
		m.Log.Warn().Msg("Could not load age rate mapping, not setting age rating")
		return "", false
	}

	var mappings models.AgeRatingMappings = m.Preference.AgeRatingMappings
	allTags := m.seriesInfo.Tags
	weights := utils.MaybeMap(allTags, func(t string) (int, bool) {
		ar, ok := mappings.GetAgeRating(t)
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
