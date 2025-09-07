package webtoon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"go.uber.org/dig"
)

func New(scope *dig.Scope) core.Downloadable {
	var wt *webtoon

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, repository Repository,
		markdownService services.MarkdownService, fs afero.Afero,
	) {
		wt = &webtoon{
			repository:      repository,
			markdownService: markdownService,
			fs:              fs,
		}

		wt.Core = core.New[Chapter, *Series](scope, "webtoon", wt)
	}))
	return wt
}

type webtoon struct {
	*core.Core[Chapter, *Series]

	repository      Repository
	markdownService services.MarkdownService
	fs              afero.Afero

	searchInfo *SearchData
}

func (w *webtoon) Title() string {
	if titleOverride, ok := w.Req.GetString(core.TitleOverride); ok {
		return titleOverride
	}

	if w.SeriesInfo == nil {
		return utils.NonEmpty(w.Req.TempTitle, w.Req.Id)
	}

	return utils.NonEmpty(w.SeriesInfo.GetTitle(), w.Req.TempTitle, w.Req.Id)
}

func (w *webtoon) Provider() models.Provider {
	return w.Req.Provider
}

func (w *webtoon) RefUrl() string {
	if w.SeriesInfo == nil {
		return ""
	}
	return w.SeriesInfo.RefUrl()
}

func (w *webtoon) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		info, err := w.repository.SeriesInfo(ctx, w.Id())
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				w.Log.Error().Err(err).Msg("error while loading webtoon info")
			}
			w.Cancel()
			return
		}

		w.SeriesInfo = info

		// TempTitle is the title we previously got from the search, just should ensure we get the correct stuff
		// WebToons search is surprisingly bad at correcting for spaces, special characters, etc...
		search, err := w.repository.Search(ctx, SearchOptions{Query: w.Req.TempTitle})
		if err != nil {
			w.Log.Error().Err(err).Msg("error while loading webtoon info")
			w.Cancel()
			return
		}

		w.searchInfo = utils.Find(search, func(data SearchData) bool {
			return data.Id == w.Id()
		})
		if w.searchInfo == nil {
			w.Log.Warn().Msg("was unable to load searchInfo, some meta-data may be off")
		}
	}()
	return out
}

func (w *webtoon) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return w.repository.LoadImages(ctx, chapter)
}

func (w *webtoon) WriteContentMetaData(ctx context.Context, chapter Chapter) error {

	if w.Req.GetBool(core.IncludeCover, true) {
		// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
		// first by most readers, and in particular, kavita.
		filePath := path.Join(w.ContentPath(chapter), "!0000 cover.jpg")
		imageUrl := func() string {
			// Kavita uses the image of the first chapter as the cover image in lists
			// We replace this with the nicer looking image. As this software is still targeting Kavita
			if w.searchInfo != nil && chapter.Number == "1" {
				return fmt.Sprintf("https://webtoon-phinf.pstatic.net%s", w.searchInfo.ThumbnailMobile)
			}
			return chapter.ImageUrl
		}()
		if err := w.DownloadAndWrite(ctx, imageUrl, filePath); err != nil {
			return err
		}
	}

	w.Log.Trace().Str("chapter", chapter.Number).Msg("writing comicinfoxml")
	return comicinfo.Save(w.fs, w.comicInfo(ctx, chapter), path.Join(w.ContentPath(chapter), "ComicInfo.xml"))
}

func (w *webtoon) comicInfo(ctx context.Context, chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = utils.NonEmpty(w.Req.GetStringOrDefault(core.TitleOverride, ""), w.Title())
	ci.Summary = w.markdownService.SanitizeHtml(w.SeriesInfo.Description)
	ci.Manga = comicinfo.MangaYes
	ci.Genre = w.SeriesInfo.Genre

	if w.searchInfo != nil {
		ci.Writer = strings.Join(w.searchInfo.AuthorNameList, ",")
		ci.AgeRating = w.searchInfo.ComicInfoRating()
		ci.Web = w.searchInfo.Url()
	}

	if chapter.Number != "" {
		ci.Number = chapter.Number
	} else {
		ci.Format = "Special"
	}

	if w.SeriesInfo.Completed {
		ci.Count = len(w.SeriesInfo.Chapters)
		w.NotifySubscriptionExhausted(ctx, fmt.Sprintf("%d Chapters", ci.Count))
	}

	return ci
}

func (w *webtoon) CustomizeRequest(req *http.Request) error {
	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")
	return nil
}
