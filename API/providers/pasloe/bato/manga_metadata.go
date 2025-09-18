package bato

import (
	"context"
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
)

func (m *manga) WriteContentMetaData(ctx context.Context, chapter Chapter) error {

	if m.Req.GetBool(core.IncludeCover, true) {
		if err := m.writeCover(ctx, chapter); err != nil {
			return err
		}
	}

	m.Log.Trace().Str("chapter", chapter.Chapter).Msg("writing comicinfoxml")
	return comicinfo.Save(m.fs, m.comicInfo(ctx, chapter), path.Join(m.ContentPath(chapter), "ComicInfo.xml"))
}

func (m *manga) writeCover(ctx context.Context, chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover"+utils.Ext(m.SeriesInfo.CoverUrl))
	return m.DownloadAndWrite(ctx, m.SeriesInfo.CoverUrl, filePath)
}

func (m *manga) comicInfo(ctx context.Context, chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = utils.NonEmpty(m.Req.GetStringOrDefault(core.TitleOverride, ""), m.SeriesInfo.Title)
	ci.AlternateSeries = m.SeriesInfo.OriginalTitle
	ci.Summary = m.SeriesInfo.Summary
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
	} else {
		ci.Format = "Special"
	}

	ci.Writer = strings.Join(utils.MaybeMap(m.SeriesInfo.Authors, func(author Author) (string, bool) {
		if author.Roles.HasRole(comicinfo.Writer) {
			return author.Name, true
		}
		return "", false
	}), ",")
	ci.Colorist = strings.Join(utils.MaybeMap(m.SeriesInfo.Authors, func(author Author) (string, bool) {
		if author.Roles.HasRole(comicinfo.Colorist) {
			return author.Name, true
		}
		return "", false
	}), ",")
	ci.Web = m.SeriesInfo.RefUrl()

	tags := utils.Map(m.SeriesInfo.Tags, core.NewStringTag)
	ci.Genre, ci.Tags = m.GetGenreAndTags(ctx, tags)
	if ar, ok := m.GetAgeRating(tags); ok {
		ci.AgeRating = ar
	}

	if count, ok := m.getCiStatus(); ok {
		ci.Count = count
		m.NotifySubscriptionExhausted(ctx, fmt.Sprintf("%d Chapters", ci.Count))
	}

	return ci

}

func (m *manga) getCiStatus() (int, bool) {
	if m.SeriesInfo.PublicationStatus != PublicationCompleted && m.SeriesInfo.BatoUploadStatus != PublicationCompleted {
		return 0, false
	}

	var highestVolume float64 = 0
	var highestChapter float64 = 0

	for _, chapter := range m.SeriesInfo.Chapters {
		if chapter.Volume != "" {
			highestVolume = math.Max(highestVolume, chapter.VolumeFloat())
		}

		if chapter.Chapter != "" {
			highestChapter = math.Max(highestChapter, chapter.ChapterFloat())
		}
	}

	if highestVolume != 0 {
		return int(highestVolume), true
	}

	if highestChapter != 0 {
		return int(highestChapter), true
	}

	return 0, false
}
