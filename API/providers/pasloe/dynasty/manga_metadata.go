package dynasty

import (
	"context"
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
	filePath := path.Join(m.ContentPath(chapter), "!0000 cover.jpg")

	if !m.hasCheckedCover {
		m.hasCheckedCover = true
		if err := m.tryReplaceCover(ctx); err != nil {
			return err
		}
	}

	if len(m.coverBytes) == 0 {
		m.Log.Trace().Str("chapter", chapter.Chapter).Msg("no cover bytes set, downloading from url")
		return m.DownloadAndWrite(ctx, m.SeriesInfo.CoverUrl, filePath)
	}

	return m.fs.WriteFile(filePath, m.coverBytes, 0755)
}

func (m *manga) tryReplaceCover(ctx context.Context) error {
	m.Log.Trace().Msg("Checking if first image of first chapter has a higher quality cover")
	firstChapter := utils.Find(m.SeriesInfo.Chapters, func(chapter Chapter) bool {
		return chapter.Chapter == "1"
	})

	if firstChapter == nil {
		return nil
	}

	images, err := m.repository.ChapterImages(ctx, firstChapter.Id)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return nil
	}

	coverBytes, err := m.Download(ctx, m.SeriesInfo.CoverUrl)
	if err != nil {
		return err
	}

	firstChapterCoverBytes, err := m.Download(ctx, images[0])
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

func (m *manga) comicInfo(ctx context.Context, chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = utils.NonEmpty(m.Req.GetStringOrDefault(core.TitleOverride, ""), m.SeriesInfo.Title)
	ci.AlternateSeries = m.SeriesInfo.AltTitle
	ci.Summary = m.markdownService.SanitizeHtml(m.SeriesInfo.Description)
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

	if count, ok := m.getCiStatus(); ok {
		ci.Count = count
	}

	ci.Writer = strings.Join(utils.Map(m.SeriesInfo.Authors, func(t Author) string {
		return t.DisplayName
	}), ",")
	ci.Web = m.SeriesInfo.RefUrl()

	tags := utils.Map(utils.FlatMapMany(chapter.Tags, m.SeriesInfo.Tags), func(t Tag) core.Tag {
		return t
	})
	ci.Genre, ci.Tags = m.GetGenreAndTags(ctx, tags)
	if ar, ok := m.GetAgeRating(tags); ok {
		ci.AgeRating = ar
	}

	return ci
}

func (m *manga) getCiStatus() (int, bool) {
	if m.SeriesInfo.Status != Completed {
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
