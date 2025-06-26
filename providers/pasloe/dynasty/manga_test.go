package dynasty

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

const (
	SailorGirlFriend = "Sailor Girlfriend"
)

type mockRepository struct {
	SearchSeriesFunc  func(ctx context.Context, options SearchOptions) ([]SearchData, error)
	SeriesInfoFunc    func(ctx context.Context, id string) (*Series, error)
	ChapterImagesFunc func(ctx context.Context, id string) ([]string, error)
}

func (m mockRepository) SearchSeries(ctx context.Context, options SearchOptions) ([]SearchData, error) {
	if m.SearchSeriesFunc == nil {
		return []SearchData{}, nil
	}
	return m.SearchSeriesFunc(ctx, options)
}

func (m mockRepository) SeriesInfo(ctx context.Context, id string) (*Series, error) {
	if m.SeriesInfoFunc == nil {
		return &Series{}, nil
	}
	return m.SeriesInfoFunc(ctx, id)
}

func (m mockRepository) ChapterImages(ctx context.Context, id string) ([]string, error) {
	if m.ChapterImagesFunc == nil {
		return []string{}, nil
	}
	return m.ChapterImagesFunc(ctx, id)
}

func tempManga(t *testing.T, req payload.DownloadRequest, w io.Writer, repo Repository) *manga {
	t.Helper()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	log := zerolog.New(w)

	c := dig.New()
	scope := c.Scope("testScope")

	tempDir := t.TempDir()
	client := mock.PasloeClient{BaseDir: tempDir}

	must(scope.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))
	must(scope.Provide(func() core.Client {
		return &client
	}))
	must(scope.Provide(utils.Identity(log)))
	must(scope.Provide(utils.Identity(menou.DefaultClient)))
	must(scope.Provide(utils.Identity(repo)))
	must(scope.Provide(utils.Identity(req)))
	must(scope.Provide(services.MarkdownServiceProvider))
	must(scope.Provide(services.ArchiveServiceProvider))
	must(scope.Provide(func() services.SignalRService { return &mock.SignalR{} }))
	must(scope.Provide(func() services.NotificationService { return &mock.Notifications{} }))
	must(scope.Provide(func() models.Preferences { return &mock.Preferences{} }))
	must(scope.Provide(func() services.TranslocoService { return &mock.Transloco{} }))
	must(scope.Provide(services.ImageServiceProvider))

	return New(scope).(*manga)
}

func req() payload.DownloadRequest {
	return payload.DownloadRequest{
		Provider:  models.DYNASTY,
		Id:        "sailor_girlfriend",
		BaseDir:   "",
		TempTitle: SailorGirlFriend,
	}
}

func chapter() Chapter {
	t := utils.MustReturn(time.Parse(RELEASEDATAFORMAT, "Jan 22 '25"))
	return Chapter{
		Id:          "sailor_girlfriend_ch4_5",
		Title:       "A Sailor's Girlfriend's Day",
		Volume:      "",
		Chapter:     "4.5",
		ReleaseDate: &t,
		Tags: []Tag{
			{
				DisplayName: "Chika x You",
				Id:          "chika_x_you",
			},
			{
				DisplayName: "College",
				Id:          "college",
			},
			{
				DisplayName: "Drunk",
				Id:          "drunk",
			},
			{
				DisplayName: "Riko x Yoshiko",
				Id:          "riko_x_yoshiko",
			},
			{
				DisplayName: "Yuri",
				Id:          "yuri",
			},
		},
	}
}

func TestManga_Title(t *testing.T) {
	var buffer bytes.Buffer

	m := tempManga(t, req(), &buffer, mockRepository{
		SeriesInfoFunc: func(ctx context.Context, id string) (*Series, error) {
			return &Series{
				Title: SailorGirlFriend,
			}, nil
		},
	})

	want := SailorGirlFriend
	if m.Title() != want {
		t.Errorf("m.Title() = %q, want %q", m.Title(), want)
	}

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	want = SailorGirlFriend
	if m.Title() != want {
		t.Errorf("m.Title() = %q, want %q", m.Title(), want)
	}
}

func TestManga_TitleInvalid(t *testing.T) {
	var buffer bytes.Buffer
	r := req()
	r.Id = "something_invalid_3456789876545678"
	m := tempManga(t, r, &buffer, mockRepository{
		SeriesInfoFunc: func(ctx context.Context, id string) (*Series, error) {
			return &Series{}, fmt.Errorf("an error occurred")
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	log := buffer.String()
	want := "error while loading series info"
	if !strings.Contains(log, want) {
		t.Errorf("m.LoadInfo() = %q, want %q", log, want)
	}

}

func TestManga_Provider(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})

	want := models.DYNASTY
	if m.Provider() != want {
		t.Errorf("m.Provider() = %q, want %q", m.Provider(), want)
	}
}

func TestManga_GetInfoBeforeLoad(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})

	got := m.GetInfo()

	want := SailorGirlFriend
	if got.Name != want {
		t.Errorf("got.Name = %q, want %q", got.Name, want)
	}

	want = ""
	if got.RefUrl != want {
		t.Errorf("got.RefUrl = %q, want %q", got.RefUrl, want)
	}

	want = "0 Chapters"
	if got.Size != want {
		t.Errorf("got.Size = %q, want %q", got.Size, want)
	}

	if got.Downloading {
		t.Errorf("got.Downloading = %v, want false", got.Downloading)
	}
}

func TestManga_GetInfoAfterLoad(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{
		SeriesInfoFunc: func(ctx context.Context, id string) (*Series, error) {
			return &Series{
				Title: SailorGirlFriend,
				Id:    "sailor_girlfriend",
			}, nil
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	got := m.GetInfo()
	want := "Sailor Girlfriend"
	if got.Name != want {
		t.Errorf("got.Name = %q, want %q", got.Name, want)
	}

	want = "https://dynasty-scans.com/series/sailor_girlfriend"
	if got.RefUrl != want {
		t.Errorf("got.RefUrl = %q, want %q", got.RefUrl, want)
	}

}

func TestManga_All(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{
		SeriesInfoFunc: func(ctx context.Context, id string) (*Series, error) {
			return &Series{
				Chapters: make([]Chapter, 5),
			}, nil
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	got := m.GetAllLoadedChapters()
	if len(got) != 5 {
		t.Errorf("len(got) = %d, want %d", len(got), 5)
	}
}

func TestManga_ContentDir(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})

	got := m.ContentFileName(chapter())
	want := SailorGirlFriend + " Ch. 0004.5"

	if got != want {
		t.Errorf("m.ContentFileName() = %q, want %q", got, want)
	}

}

func TestManga_ContentPath(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})

	got := m.ContentPath(chapter())
	want := path.Join(m.Client.GetBaseDir(), fmt.Sprintf("%s/%s Ch. 0004.5", SailorGirlFriend, SailorGirlFriend))
	if got != want {
		t.Errorf("m.ContentPath() = %q, want %q", got, want)
	}
}

func TestManga_ContentLogger(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})

	log := m.ContentLogger(chapter())

	log.Info().Msg("a")

	c := chapter()
	c.Volume = "1"
	log = m.ContentLogger(c)
	log.Info().Msg("b")

	want1 := "{\"level\":\"info\",\"handler\":\"dynasty-manga\",\"id\":\"sailor_girlfriend\",\"chapterId\":\"sailor_girlfriend_ch4_5\",\"chapter\":\"4.5\",\"title\":\"A Sailor's Girlfriend's Day\",\"message\":\"a\"}"
	want2 := "{\"level\":\"info\",\"handler\":\"dynasty-manga\",\"id\":\"sailor_girlfriend\",\"chapterId\":\"sailor_girlfriend_ch4_5\",\"chapter\":\"4.5\",\"title\":\"A Sailor's Girlfriend's Day\",\"volume\":\"1\",\"message\":\"b\"}"

	s := buffer.String()
	if !strings.Contains(s, want1) {
		t.Errorf("m.Content() = %q, want %q", s, want1)
	}

	if !strings.Contains(s, want2) {
		t.Errorf("m.Content() = %q, want %q", s, want2)
	}
}

func TestManga_ContentUrls(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{
		ChapterImagesFunc: func(ctx context.Context, id string) ([]string, error) {
			return make([]string, 19), nil
		},
	})

	urls, err := m.ContentUrls(t.Context(), chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != 19 {
		t.Errorf("len(urls) = %d, want %d", len(urls), 19)
	}

}

func TestManga_ContentRegex(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})

	if _, ok := m.IsContent("Not a Valid Chapter"); ok {
		t.Error("m.ContentRegex().MatchString() returned true")
	}

	if _, ok := m.IsContent("Sailor Girlfriend Ch. 0004.5.cbz"); !ok {
		t.Error("m.ContentRegex().MatchString() returned false")
	}

	if _, ok := m.IsContent("Shiawase Trimming OneShot Manga Time Kirara 20th Anniversary Special Collaboration: Stardust Telepath x Shiawase Trimming.cbz"); !ok {
		t.Error("m.ContentRegex().MatchString() returned false")
	}
}

func TestManga_ShouldDownload(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer, &mockRepository{})
	m.ExistingContent = []core.Content{
		{
			Name: "Sailor Girlfriend Ch. 0001.cbz",
		},
		{
			Name: "Sailor Girlfriend Ch. 0002.cbz",
		},
	}

	/*select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}*/

	chapter1 := chapter()
	chapter1.Chapter = "1"
	got := m.ShouldDownload(chapter1)
	if got != false {
		t.Errorf("m.ShouldDownload() = %t, want false", got)
	}

	got = m.ShouldDownload(chapter())
	if got != true {
		t.Errorf("m.ShouldDownload() = %t, want true", got)
	}
}

func TestCoverReplace(t *testing.T) {
	time.Sleep(1 * time.Second)
	m := tempManga(t, payload.DownloadRequest{
		Provider:  models.DYNASTY,
		Id:        "a_story_about_buying_a_classmate_once_a_week_5000_yen_for_an_excuse_to_spend_time_together",
		BaseDir:   "Manga",
		TempTitle: "A Story About Buying a Classmate Once a Week",
		DownloadMetadata: models.DownloadRequestMetadata{
			StartImmediately: true,
		},
		IsSubscription: false,
	}, io.Discard, &mockRepository{
		SeriesInfoFunc: func(ctx context.Context, id string) (*Series, error) {
			return &Series{
				CoverUrl: "https://dynasty-scans.com/system/tag_contents_covers/000/023/149/medium/00.jpg?1718830047",
				Chapters: []Chapter{
					{
						Chapter: "1",
					},
				},
			}, nil
		},
		ChapterImagesFunc: func(ctx context.Context, id string) ([]string, error) {
			return []string{"https://dynasty-scans.com/system/releases/000/042/230/00.webp"}, nil
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	if err := m.tryReplaceCover(); err != nil {
		t.Fatal(err)
	}

	originalCover, err := m.Download(m.SeriesInfo.CoverUrl, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(originalCover) == len(m.coverBytes) {
		t.Error("Cover should have been replaced, but wasn't")
	}
}

func TestCoverNoReplace(t *testing.T) {
	time.Sleep(1 * time.Second)
	m := tempManga(t, payload.DownloadRequest{
		Provider:  models.DYNASTY,
		Id:        "the_blue_star_on_that_day_ano_koro_no_aoi_hoshi",
		BaseDir:   "Manga",
		TempTitle: "The Blue Star on That Day",
		DownloadMetadata: models.DownloadRequestMetadata{
			StartImmediately: true,
		},
		IsSubscription: false,
	}, io.Discard, &mockRepository{
		SeriesInfoFunc: func(ctx context.Context, id string) (*Series, error) {
			return &Series{
				CoverUrl: "https://dynasty-scans.com/system/tag_contents_covers/000/013/315/medium/cover.jpg?1663754407",
				Chapters: []Chapter{
					{
						Chapter: "1",
					},
				},
			}, nil
		},
		ChapterImagesFunc: func(ctx context.Context, id string) ([]string, error) {
			return []string{"https://dynasty-scans.com/system/releases/000/026/104/001.webp"}, nil
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	if err := m.tryReplaceCover(); err != nil {
		if strings.Contains(err.Error(), "503") {
			t.Skipf("Skipping due to 503 error: %v", err)
		}
		t.Fatal(err)
	}

	originalCover, err := m.Download(m.SeriesInfo.CoverUrl, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(originalCover) != len(m.coverBytes) {
		t.Error("Cover should not have been replaced, but was")
	}
}
