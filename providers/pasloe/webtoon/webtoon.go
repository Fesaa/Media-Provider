package webtoon

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"
)

func NewWebToon(scope *dig.Scope) core.Downloadable {
	var wt *webtoon

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, httpClient *menou.Client,
		repository Repository, markdownService services.MarkdownService,
		fs afero.Afero,
	) {
		wt = &webtoon{
			id:              req.Id,
			httpClient:      httpClient,
			repository:      repository,
			markdownService: markdownService,
			fs:              fs,
		}

		wt.Core = core.New[Chapter](scope, "webtoon", wt)
	}))
	return wt
}

type webtoon struct {
	httpClient      *menou.Client
	repository      Repository
	markdownService services.MarkdownService
	fs              afero.Afero

	*core.Core[Chapter]
	id string

	searchInfo *SearchData
	info       *Series
}

func (w *webtoon) Title() string {
	if w.searchInfo != nil {
		return w.searchInfo.Name
	}
	if w.info != nil {
		return w.info.Name
	}

	if w.Req.TempTitle != "" {
		return w.Req.TempTitle
	}

	return w.id
}

func (w *webtoon) Provider() models.Provider {
	return w.Req.Provider
}

func (w *webtoon) RefUrl() string {
	if w.searchInfo != nil {
		return w.searchInfo.Url()
	}

	return ""
}

func (w *webtoon) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		info, err := w.repository.SeriesInfo(ctx, w.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				w.Log.Error().Err(err).Msg("error while loading webtoon info")
			}
			w.Cancel()
			return
		}

		w.info = info

		// TempTitle is the title we previously got from the search, just should ensure we get the correct stuff
		// WebToons search is surprisingly bad at correcting for spaces, special characters, etc...
		search, err := w.repository.Search(ctx, SearchOptions{Query: w.Req.TempTitle})
		if err != nil {
			w.Log.Error().Err(err).Msg("error while loading webtoon info")
			w.Cancel()
			return
		}

		w.searchInfo = utils.Find(search, func(data SearchData) bool {
			return data.Id == w.id
		})
		if w.searchInfo == nil {
			w.Log.Warn().Msg("was unable to load searchInfo, some meta-data may be off")
		}
	}()
	return out
}

func (w *webtoon) All() []Chapter {
	return w.info.Chapters
}

func (w *webtoon) ContentList() []payload.ListContentData {
	if w.info == nil {
		return nil
	}

	return utils.Map(w.info.Chapters, func(chapter Chapter) payload.ListContentData {
		return payload.ListContentData{
			SubContentId: chapter.Number,
			Selected:     len(w.ToDownloadUserSelected) == 0 || slices.Contains(w.ToDownloadUserSelected, chapter.Number),
			Label:        fmt.Sprintf("%s #%s - %s", w.info.Name, chapter.Number, chapter.Title),
		}
	})
}

func (w *webtoon) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return w.repository.LoadImages(ctx, chapter)
}

func (w *webtoon) WriteContentMetaData(chapter Chapter) error {

	if w.Req.GetBool(IncludeCover, true) {
		// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
		// first by most readers, and in particular, kavita.
		filePath := path.Join(w.ContentPath(chapter), "!0000 cover.jpg")
		imageUrl := func() string {
			// Kavita uses the image of the first chapter as the cover image in lists
			// We replace this with the nicer looking image. As this software is still targeting Kavita
			if w.searchInfo != nil && chapter.Number == "1" {
				return webToonUrl(w.searchInfo.ThumbnailMobile)
			}
			return chapter.ImageUrl
		}()
		if err := w.DownloadAndWrite(imageUrl, filePath); err != nil {
			return err
		}
	}

	w.Log.Trace().Str("chapter", chapter.Number).Msg("writing comicinfoxml")
	return comicinfo.Save(w.fs, w.comicInfo(chapter), path.Join(w.ContentPath(chapter), "ComicInfo.xml"))
}

func (w *webtoon) comicInfo(chapter Chapter) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = w.Title()
	ci.Summary = w.markdownService.SanitizeHtml(w.info.Description)
	ci.Manga = comicinfo.MangaYes
	ci.Genre = w.info.Genre

	if w.searchInfo != nil {
		ci.Writer = strings.Join(w.searchInfo.AuthorNameList, ",")
		ci.AgeRating = w.searchInfo.ComicInfoRating()
		ci.Web = w.searchInfo.Url()
	}

	if chapter.Number != "" {
		ci.Number = chapter.Number
	}

	if w.info.Completed {
		ci.Count = len(w.info.Chapters)
	}

	return ci
}

var chapterRegex = regexp.MustCompile(".* Ch\\. (\\d+).cbz")

func (w *webtoon) IsContent(name string) bool {
	return chapterRegex.MatchString(name)
}

func (w *webtoon) ShouldDownload(chapter Chapter) bool {
	_, ok := w.GetContentByName(w.ContentDir(chapter) + ".cbz")
	return !ok
}

func (w *webtoon) CustomizeRequest(req *http.Request) error {
	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")
	return nil
}

func webToonUrl(s string) string {
	return fmt.Sprintf("https://webtoon-phinf.pstatic.net%s", s)
}
