package bato

import (
	"context"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
	"path"
	"strconv"
	"strings"
)

func (m *manga) WriteContentMetaData(ctx context.Context, chapter Chapter) error {

	if m.Req.GetBool(core.IncludeCover, true) {
		if err := m.writeCover(ctx, chapter); err != nil {
			return err
		}
	}

	m.Log.Trace().Str("chapter", chapter.Chapter).Msg("writing comicinfoxml")
	return comicinfo.Save(m.fs, m.comicInfo(chapter), path.Join(m.ContentPath(chapter), "ComicInfo.xml"))
}

func (m *manga) writeCover(ctx context.Context, chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover"+utils.Ext(m.SeriesInfo.CoverUrl))
	return m.DownloadAndWrite(ctx, m.SeriesInfo.CoverUrl, filePath)
}

func (m *manga) comicInfo(chapter Chapter) *comicinfo.ComicInfo {
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
	ci.Genre, ci.Tags = m.GetGenreAndTags(tags)
	if ar, ok := m.GetAgeRating(tags); ok {
		ci.AgeRating = ar
	}

	if m.SeriesInfo.PublicationStatus == PublicationCompleted && m.SeriesInfo.BatoUploadStatus == PublicationCompleted {
		ci.Count = m.ChapterCount()
	}

	return ci

}

func (m *manga) ChapterCount() int {
	c := 0
	for _, ch := range m.SeriesInfo.Chapters {
		if v, err := strconv.Atoi(ch.Chapter); err == nil {
			c = max(c, v)
		}
	}

	return c
}
