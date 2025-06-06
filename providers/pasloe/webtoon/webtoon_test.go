package webtoon

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/utils/mock"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"io"
	"path"
	"strings"
	"testing"
	"time"
)

func req() payload.DownloadRequest {
	return payload.DownloadRequest{
		Provider:  models.WEBTOON,
		Id:        WebToonID,
		BaseDir:   "Manga",
		TempTitle: WebToonName,
	}
}

func chapter() Chapter {
	return Chapter{
		Url:      "https://www.webtoons.com/en/romance/night-owls-and-summer-skies/episode-8/viewer?title_no=4747&episode_no=8",
		ImageUrl: "https://webtoon-phinf.pstatic.net/20220920_282/1663629235464jKwS9_PNG/thumb_1663613574742474787.png?type=q90",
		Title:    "Episode 8",
		Number:   "8",
		Date:     "Oct 26, 2022",
	}
}

func series() *Series {
	return &Series{
		Id:          "4747",
		Name:        WebToonName,
		Authors:     []string{"TIKKLIL", "Rebecca Sullivan"},
		Description: "Despite her tough exterior, seventeen year-old Emma Lane has never been the outdoorsy type. So when her mother unceremoniously dumps her at Camp Mapplewood for the summer, she’s determined to get kicked out fast. However, when she draws the attention of Vivian Black, a mysterious and gorgeous assistant counselor, she discovers that there may be more to this camp than mean girls and mosquitos. There might even be love.",
		Genre:       "romance",
		Completed:   true,
		Chapters:    []Chapter{chapter()},
	}
}

func tempWebtoon(t *testing.T, w io.Writer) *webtoon {
	t.Helper()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	client := mock.PasloeClient{BaseDir: t.TempDir()}

	cont := dig.New()
	scope := cont.Scope("tempWebtoon")

	must(scope.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))
	must(scope.Provide(func() core.Client {
		return client
	}))
	must(scope.Provide(utils.Identity(menou.DefaultClient)))
	must(scope.Provide(utils.Identity(zerolog.New(w))))
	must(scope.Provide(utils.Identity(req())))
	must(scope.Provide(NewRepository))
	must(scope.Provide(services.MarkdownServiceProvider))
	must(scope.Provide(func() services.SignalRService { return &mock.SignalR{} }))
	must(scope.Provide(func() services.NotificationService { return &mock.Notifications{} }))
	must(scope.Provide(func() models.Preferences { return &mock.Preferences{} }))
	must(scope.Provide(func() services.TranslocoService { return &mock.Transloco{} }))

	web := NewWebToon(scope)
	return web.(*webtoon)
}

type mockRepository struct {
	searchRes []SearchData
	searchErr error
	imageRes  []string
	imageErr  error
	infoRes   *Series
	infoErr   error
}

func (m *mockRepository) Search(ctx context.Context, options SearchOptions) ([]SearchData, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.searchRes, nil
}

func (m *mockRepository) LoadImages(ctx context.Context, chapter Chapter) ([]string, error) {
	if m.imageErr != nil {
		return nil, m.imageErr
	}
	return m.imageRes, nil
}

func (m *mockRepository) SeriesInfo(ctx context.Context, id string) (*Series, error) {
	if m.infoErr != nil {
		return nil, m.infoErr
	}
	return m.infoRes, nil
}

func TestWebtoon_Title(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	want := req().TempTitle
	got := wt.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	select {
	case <-wt.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for info")
	}

	want = WebToonName
	got = wt.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	wt.searchInfo = nil
	got = wt.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	wt.info = nil
	wt.Req.TempTitle = ""

	want = "4747"
	got = wt.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

}

func TestWebtoon_Provider(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	want := models.WEBTOON
	if got := wt.Provider(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWebtoon_LoadInfo(t *testing.T) {
	var buffer bytes.Buffer
	wt := tempWebtoon(t, &buffer)

	repo := mockRepository{}
	wt.repository = &repo

	repo.infoErr = fmt.Errorf("error")

	select {
	case <-wt.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for info")
	}

	log := buffer.String()
	want := "error while loading webtoon info"
	if !strings.Contains(log, want) {
		t.Errorf("got %q, want %q", log, want)
	}
	buffer.Reset()

	repo.infoErr = nil
	repo.searchErr = fmt.Errorf("error")
	select {
	case <-wt.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for info")
	}

	log = buffer.String()
	if !strings.Contains(log, want) {
		t.Errorf("got %q, want %q", log, want)
	}
	buffer.Reset()

	repo.searchErr = nil
	repo.searchRes = []SearchData{
		{
			Id:   "0",
			Name: "An other series",
		},
	}

	select {
	case <-wt.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for info")
	}

	log = buffer.String()
	want = "was unable to load searchInfo, some meta-data may be off"
	if !strings.Contains(log, want) {
		t.Errorf("got %q, want %q", log, want)
	}
}

func TestWebtoon_All(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	repo := mockRepository{}
	wt.repository = &repo

	repo.infoRes = &Series{
		Chapters: []Chapter{
			{
				Title:  "Chapter 1",
				Number: "1",
			},
			{
				Title:  "Chapter 2",
				Number: "2",
			},
		},
	}

	select {
	case <-wt.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for info")
	}

	want := 2
	got := wt.All()
	if len(got) != want {
		t.Errorf("got %d, want %d", len(got), want)
	}
}

func TestWebtoon_ContentDir(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	want := WebToonName + " Ch. 0008"
	got := wt.ContentDir(chapter())
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWebtoon_ContentPath(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	want := path.Join(WebToonName, WebToonName+" Ch. 0008")
	got := wt.ContentPath(chapter())

	if !strings.HasSuffix(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWebtoon_ContentKey(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)
	want := "8"
	got := wt.ContentKey(chapter())
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWebtoon_ContentUrls(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)
	mock := mockRepository{}
	wt.repository = &mock

	mock.imageRes = []string{"image1", "image2"}

	want := 2
	got, err := wt.ContentUrls(t.Context(), chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != want {
		t.Errorf("got %d, want %d", len(got), want)
	}

	mock.imageErr = fmt.Errorf("error")
	got, err = wt.ContentUrls(t.Context(), chapter())
	if err == nil {
		t.Errorf("got %v, want error", got)
	}

}

func TestWebtoon_DownloadContent(t *testing.T) {
	var buffer bytes.Buffer
	w := tempWebtoon(t, &buffer)

	urls, err := w.ContentUrls(t.Context(), chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) == 0 {
		t.Fatal("len(urls) = 0, want > 0")
	}

	if err = w.fs.MkdirAll(path.Join(w.ContentPath(chapter())), 0755); err != nil {
		t.Fatal(err)
	}

	if err = w.DownloadContent(1, chapter(), urls[0]); err != nil {
		t.Fatal(err)
	}

	filePath := path.Join(w.ContentPath(chapter()), fmt.Sprintf("page %s.jpg", utils.PadInt(1, 4)))
	_, err = w.fs.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWebtoon_ContentRegex(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	type test struct {
		name  string
		s     string
		match bool
	}

	tests := []test{
		{
			name:  "Match",
			s:     WebToonName + " Ch. 8.cbz",
			match: true,
		},
		{
			name:  "No Match",
			s:     "DFGHJKL",
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wt.IsContent(tt.s); got != tt.match {
				t.Errorf("got %v, want %v", got, tt.match)
			}
		})
	}
}
