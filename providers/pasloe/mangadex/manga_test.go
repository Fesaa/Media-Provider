package mangadex

import (
	"bytes"
	"context"
	"errors"
	"github.com/Fesaa/Media-Provider/comicinfo"
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
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

const (
	waitTimeOut = 30 * time.Second
)

type mockRepository struct {
	GetMangaFunc         func(ctx context.Context, id string) (*GetMangaResponse, error)
	SearchMangaFunc      func(ctx context.Context, options SearchOptions) (*MangaSearchResponse, error)
	GetChaptersFunc      func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error)
	GetChapterImagesFunc func(ctx context.Context, id string) (*ChapterImageSearchResponse, error)
	GetCoverImagesFunc   func(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error)
}

func (m mockRepository) GetManga(ctx context.Context, id string) (*GetMangaResponse, error) {
	if m.GetMangaFunc == nil {
		return &GetMangaResponse{}, nil
	}
	return m.GetMangaFunc(ctx, id)
}

func (m mockRepository) SearchManga(ctx context.Context, options SearchOptions) (*MangaSearchResponse, error) {
	if m.SearchMangaFunc == nil {
		return &MangaSearchResponse{}, nil
	}
	return m.SearchMangaFunc(ctx, options)
}

func (m mockRepository) GetChapters(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
	if m.GetChaptersFunc == nil {
		return &ChapterSearchResponse{}, nil
	}
	return m.GetChaptersFunc(ctx, id, offset...)
}

func (m mockRepository) GetChapterImages(ctx context.Context, id string) (*ChapterImageSearchResponse, error) {
	if m.GetChapterImagesFunc == nil {
		return &ChapterImageSearchResponse{}, nil
	}
	return m.GetChapterImagesFunc(ctx, id)
}

func (m mockRepository) GetCoverImages(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error) {
	if m.GetCoverImagesFunc == nil {
		return &MangaCoverResponse{}, nil
	}
	return m.GetCoverImagesFunc(ctx, id, offset...)
}

func tempManga(t *testing.T, req payload.DownloadRequest, w io.Writer, repo Repository) *manga {
	t.Helper()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	log := zerolog.New(w).Level(zerolog.TraceLevel)

	c := dig.New()
	scope := c.Scope("testScope")

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	must(fs.Mkdir("/data", 0755))
	client := mock.PasloeClient{BaseDir: "/data"}

	must(scope.Provide(utils.Identity(fs)))
	must(scope.Provide(func() core.Client {
		return &client
	}))
	must(scope.Provide(utils.Identity(log)))
	must(scope.Provide(utils.Identity(menou.DefaultClient)))
	must(scope.Provide(utils.Identity(repo)))
	must(scope.Provide(utils.Identity(req)))
	must(scope.Provide(services.MarkdownServiceProvider))
	must(scope.Provide(func() services.SignalRService { return &mock.SignalR{} }))
	must(scope.Provide(func() services.NotificationService { return &mock.Notifications{} }))
	must(scope.Provide(func() models.Preferences { return &mock.Preferences{} }))
	must(scope.Provide(func() services.TranslocoService { return &mock.Transloco{} }))
	must(scope.Provide(func() services.CacheService { return &mock.Cache{} }))
	must(scope.Provide(services.ImageServiceProvider))
	must(scope.Provide(services.ArchiveServiceProvider))

	return NewManga(scope).(*manga)
}

func req() payload.DownloadRequest {
	return payload.DownloadRequest{
		Provider:  models.MANGADEX,
		Id:        RainbowsAfterStormsID,
		BaseDir:   "",
		TempTitle: RainbowsAfterStorms,
	}
}

func chapter() ChapterSearchData {
	return ChapterSearchData{
		Id:   RainbowsAfterStormsLastChapterID,
		Type: "chapter",
		Attributes: ChapterAttributes{
			Volume:             "13",
			Chapter:            "162",
			Title:              "My Lover",
			TranslatedLanguage: "en",
			Pages:              22,
			PublishedAt:        "2024-09-29T15:41:17+00:00",
		},
		Relationships: nil,
	}
}

func mangaResp() *MangaSearchData {
	return &MangaSearchData{
		Attributes: MangaAttributes{
			Title: map[string]string{
				"en": "Rainbows After Storms",
			},
			AltTitles: []map[string]string{
				{
					"en": "Flowers After Storms",
				},
			},
			Tags: []TagData{
				{
					Attributes: TagAttributes{
						Name:  map[string]string{"en": "A Genre"},
						Group: genreTag,
					},
				},
				{
					Attributes: TagAttributes{
						Name:  map[string]string{"en": "Not a Genre"},
						Group: "not a genre",
					},
				},
				{
					Attributes: TagAttributes{
						Name:  map[string]string{"cn": "Non English"},
						Group: genreTag,
					},
				},
				{
					Attributes: TagAttributes{
						Name:  map[string]string{"en": "Blacklisted Genre"},
						Group: genreTag,
					},
				},
				{
					Attributes: TagAttributes{
						Name:  map[string]string{"en": "Blacklisted Tag"},
						Group: "tag",
					},
				},
				{
					Id: "ABC",
					Attributes: TagAttributes{
						Name:  map[string]string{"en": "Something random"},
						Group: genreTag,
					},
				},
			},
		},
	}
}

func TestManga_Title(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{
		GetMangaFunc: func(ctx context.Context, id string) (*GetMangaResponse, error) {
			return &GetMangaResponse{
				Data: MangaSearchData{
					Attributes: MangaAttributes{
						Title: map[string]string{
							"en": RainbowsAfterStorms,
						},
					},
				},
			}, nil
		},
	})

	want := RainbowsAfterStorms
	got := m.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}

	want = RainbowsAfterStorms
	got = m.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestManga_LoadInfoBadId(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf, &mockRepository{
		GetMangaFunc: func(ctx context.Context, id string) (*GetMangaResponse, error) {
			return nil, errors.New("some error")
		},
	})

	m.id = "DFGHJKJHGFDGHJKHGFGHJ"
	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}

	log := buf.String()
	if !strings.Contains(log, "error while loading manga info") {
		t.Errorf("got %q, want 'error while loading manga info'", log)
	}

}

//nolint:funlen
func TestManga_LoadInfoFoundAll(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf, &mockRepository{
		GetMangaFunc: func(ctx context.Context, id string) (*GetMangaResponse, error) {
			return &GetMangaResponse{
				Data: MangaSearchData{
					Attributes: MangaAttributes{
						LastChapter: "8",
						LastVolume:  "2",
					},
				},
			}, nil
		},
		GetChaptersFunc: func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
			return &ChapterSearchResponse{
				Data: []ChapterSearchData{
					{
						Attributes: ChapterAttributes{
							TranslatedLanguage: "en",
							Volume:             "1",
							Chapter:            "2",
						},
					},
					{
						Attributes: ChapterAttributes{
							TranslatedLanguage: "en",
							Volume:             "2",
							Chapter:            "8",
						},
					},
				},
			}, nil
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}

	if !m.foundLastVolume {
		t.Error("expected manga to have last chapter")
	}

	if !m.foundLastChapter {
		t.Error("expected manga to have last chapter")
	}

	got := m.lastFoundVolume
	want := 2

	if got != want {
		t.Errorf("lastFoundVolume got %d, want %d", got, want)
	}

	got = m.lastFoundChapter
	want = 8
	if got != want {
		t.Errorf("lastFoundChapter got %d, want %d", got, want)
	}
}

//nolint:funlen
func TestManga_LoadInfoErrors(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf, &mockRepository{
		GetMangaFunc: func(ctx context.Context, id string) (*GetMangaResponse, error) {
			return nil, errors.New("some error")
		},
		GetChaptersFunc: func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
			return nil, errors.New("some error")
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}

	log := buf.String()
	if !strings.Contains(log, "error while loading manga info") {
		t.Errorf("got %q, want 'error while loading manga info'", log)
	}
	buf.Reset()

	m.repository.(*mockRepository).GetMangaFunc = func(ctx context.Context, id string) (*GetMangaResponse, error) {
		return &GetMangaResponse{}, nil
	}
	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}

	log = buf.String()
	if !strings.Contains(log, "error while loading chapter info") {
		t.Errorf("got %q, want 'error while loading chapter info'", log)
	}
	buf.Reset()

	m.repository.(*mockRepository).GetChaptersFunc = func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
		return &ChapterSearchResponse{}, nil
	}

	m.repository.(*mockRepository).GetCoverImagesFunc = func(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error) {
		return &MangaCoverResponse{}, errors.New("some error")
	}

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}
	log = buf.String()
	if !strings.Contains(log, "error while loading manga coverFactory") {
		t.Errorf("got %q, want 'error while loading manga coverFactory'", log)
	}
	buf.Reset()

	m.repository.(*mockRepository).GetChaptersFunc = func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
		return &ChapterSearchResponse{
			Data: []ChapterSearchData{
				{Attributes: ChapterAttributes{
					TranslatedLanguage: "en",
					Volume:             "NotANumber",
					Chapter:            "1", // Needed so it doesn't get picked up as a OneShot, and skipped
				}},
			},
		}, nil
	}
	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Error("timed out waiting for manga title")
	}

	log = buf.String()
	if !strings.Contains(log, "not adding chapter, as Volume string isn't an int") {
		t.Errorf("got %q, want 'not adding chapter, as Volume isn't an int'", log)
	}

}

func TestManga_Provider(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})
	if m.Provider() != models.MANGADEX {
		t.Errorf("got %q, want %q", m.Provider(), models.MANGADEX)
	}
}

func TestManga_All(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{
		GetChaptersFunc: func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
			return &ChapterSearchResponse{
				Data: []ChapterSearchData{
					{Attributes: ChapterAttributes{
						TranslatedLanguage: "en",
						Chapter:            "1", // Needed so it doesn't get picked up as a OneShot, and skipped
					}},
				},
			}, nil
		},
	})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Fatal("timed out waiting for manga title")
	}

	got := m.All()
	want := 1

	if len(got) != want {
		t.Errorf("got %d, want %d", len(got), want)
	}

}

func TestManga_ContentDir(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})
	m.info = mangaResp()

	got := m.ContentDir(chapter())
	want := "Rainbows After Storms Ch. 0162"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestManga_ContentDirBadChapter(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf, &mockRepository{})
	m.info = mangaResp()

	chpt := chapter()
	chpt.Attributes.Chapter = "NotAFloat"
	got := m.ContentDir(chpt)
	want := "Rainbows After Storms Ch. NotAFloat"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	log := buf.String()
	if !strings.Contains(log, "unable to parse chpt number, not padding") {
		t.Errorf("got %q, want 'unable to parse chpt number, not padding'", log)
	}
	buf.Reset()

	chpt.Attributes.Chapter = ""
	got = m.ContentDir(chpt)
	want = "Rainbows After Storms OneShot My Lover"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	log = buf.String()
	if strings.Contains(log, "unable to parse chpt number, not padding") {
		t.Errorf("got %q, didn't want 'unable to parse chpt number, not padding'", log)
	}
	buf.Reset()

}

func TestManga_ContentPath(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})
	m.info = mangaResp()

	got := m.ContentPath(chapter())
	want := "Rainbows After Storms/Rainbows After Storms Vol. 13/" + m.ContentDir(chapter())
	if !strings.HasSuffix(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}

	chpt := chapter()
	chpt.Attributes.Volume = ""
	got = m.ContentPath(chpt)
	want = "Rainbows After Storms/" + m.ContentDir(chpt)
	if !strings.HasSuffix(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestManga_ContentKey(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})

	got := m.ContentKey(chapter())
	want := RainbowsAfterStormsLastChapterID
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestManga_ContentLogger(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf, &mockRepository{})

	log := m.ContentLogger(chapter())
	log.Info().Msg("a")

	want := "{\"level\":\"info\",\"handler\":\"mangadex\",\"id\":\"bc86a871-ddc5-4e42-812a-ccd38101d82e\",\"chapterId\":\"7d327897-5903-4fa1-92d7-f01c3c686a36\",\"chapter\":\"162\",\"volume\":\"13\",\"title\":\"My Lover\",\"message\":\"a\"}"
	out := buf.String()
	if !strings.Contains(out, want) {
		t.Errorf("got %s, want %s", buf.String(), want)
	}
	buf.Reset()

	chpt := chapter()
	chpt.Attributes.Volume = ""
	log = m.ContentLogger(chpt)
	log.Info().Msg("b")

	out = buf.String()
	want = "{\"level\":\"info\",\"handler\":\"mangadex\",\"id\":\"bc86a871-ddc5-4e42-812a-ccd38101d82e\",\"chapterId\":\"7d327897-5903-4fa1-92d7-f01c3c686a36\",\"chapter\":\"162\",\"title\":\"My Lover\",\"message\":\"b\"}"
	if !strings.Contains(out, want) {
		t.Errorf("got %s, want %s", buf.String(), want)
	}
	buf.Reset()

}

func TestManga_ContentUrls(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{
		GetChapterImagesFunc: func(ctx context.Context, id string) (*ChapterImageSearchResponse, error) {
			return &ChapterImageSearchResponse{
				Chapter: ChapterInfo{
					Data: []string{"1", "2", "3"},
				},
			}, nil
		},
	})

	urls, err := m.ContentUrls(t.Context(), chapter())
	if err != nil {
		t.Fatal(err)
	}

	want := 3
	if len(urls) != want {
		t.Errorf("got %d, want %d", len(urls), want)
	}

	m.repository.(*mockRepository).GetChapterImagesFunc = func(ctx context.Context, id string) (*ChapterImageSearchResponse, error) {
		return nil, errors.New("error")
	}
	_, err = m.ContentUrls(t.Context(), chapter())
	if err == nil {
		t.Errorf("got %v, want error", urls)
	}

}

func TestManga_ContentRegex(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})

	type test struct {
		name string
		s    string
		want bool
	}

	tests := []test{
		{
			name: "Test Volume",
			s:    RainbowsAfterStorms + " Vol. 7.cbz",
			want: true,
		},
		{
			name: "Test Chapter",
			s:    RainbowsAfterStorms + " Ch. 7.cbz",
			want: true,
		},
		{
			name: "Test Random",
			s:    "DFGHJK",
			want: false,
		},
		{
			name: "Test OneShot",
			s:    RainbowsAfterStorms + " OneShot Shopping With Friends.cbz",
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if m.IsContent(tc.s) != tc.want {
				t.Errorf("got %v, want %v for %s", !tc.want, tc.want, tc.s)
			}
		})
	}

}

//nolint:funlen,gocognit
func TestManga_ShouldDownload(t *testing.T) {

	type test struct {
		name          string
		contentOnDisk []core.Content
		chapter       func() ChapterSearchData
		command       func(*testing.T, *manga)
		want          bool
		logInclude    string
		after         func(*testing.T, *manga)
	}

	tests := []test{
		{
			name:          "New Download",
			contentOnDisk: []core.Content{},
			chapter:       chapter,
			want:          true,
		},
		{
			name: "Volume on disk",
			contentOnDisk: []core.Content{
				{Name: RainbowsAfterStorms + " Vol. 13.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13.cbz")},
			},
			chapter: chapter,
			want:    false,
		},
		{
			name: "Chapter on disk, no volume",
			contentOnDisk: []core.Content{
				{Name: RainbowsAfterStorms + " Ch. 0162.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 0162.cbz")},
			},
			chapter: func() ChapterSearchData {
				ch := chapter()
				ch.Attributes.Volume = ""
				return ch
			},
			want: false,
		},
		{
			name: "Chapter on disk, fail volume check",
			contentOnDisk: []core.Content{
				{Name: RainbowsAfterStorms + " Ch. 0162.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13", RainbowsAfterStorms+" Ch. 0162.cbz")},
			},
			chapter:    chapter,
			command:    nil,
			want:       false,
			logInclude: "unable to read comic info in zip",
			after:      nil,
		},
		{
			name: "Chapter on disk, same volume on disk",
			contentOnDisk: []core.Content{
				{Name: RainbowsAfterStorms + " Ch. 0162.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13", RainbowsAfterStorms+" Ch. 0162.cbz")},
			},
			chapter: chapter,
			command: func(t *testing.T, m *manga) {
				t.Helper()
				ds := services.DirectoryServiceProvider(zerolog.Nop(), m.fs)

				fullpath := path.Join(m.Client.GetBaseDir(), RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13", RainbowsAfterStorms+" Ch. 0162")
				ci := comicinfo.NewComicInfo()
				ci.Volume = 13
				if err := m.fs.MkdirAll(fullpath, 0755); err != nil {
					t.Fatal(err)
				}
				if err := comicinfo.Save(m.fs, ci, path.Join(fullpath, "comicinfo.xml")); err != nil {
					t.Fatal(err)
				}

				if err := ds.ZipFolder(fullpath, fullpath+".cbz"); err != nil {
					t.Fatal(err)
				}
			},
			want:       false,
			logInclude: "Volume on disk matches, not replacing",
			after:      nil,
		},
		{
			name: "Chapter on disk, replacing loose chapter",
			contentOnDisk: []core.Content{
				{Name: RainbowsAfterStorms + " Ch. 0162.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 0162.cbz")},
			},
			chapter: chapter,
			command: func(t *testing.T, m *manga) {
				t.Helper()
				ds := services.DirectoryServiceProvider(zerolog.Nop(), m.fs)

				fullpath := path.Join(m.Client.GetBaseDir(), RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 0162")
				ci := comicinfo.NewComicInfo()
				if err := m.fs.MkdirAll(fullpath, 0755); err != nil {
					t.Fatal(err)
				}
				if err := comicinfo.Save(m.fs, ci, path.Join(fullpath, "comicinfo.xml")); err != nil {
					t.Fatal(err)
				}

				if err := ds.ZipFolder(fullpath, fullpath+".cbz"); err != nil {
					t.Fatal(err)
				}
			},
			want:       true,
			logInclude: "Loose chapter has been assigned to a volume, replacing",
			after: func(t *testing.T, m *manga) {
				t.Helper()
				fullpath := path.Join(m.Client.GetBaseDir(), RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 162.cbz")
				_, err := m.fs.Stat(fullpath)
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("expected %s to not exist", fullpath)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			m := tempManga(t, req(), &buffer, &mockRepository{})
			m.ExistingContent = tc.contentOnDisk
			m.info = mangaResp()

			if tc.command != nil {
				tc.command(t, m)
			}

			got := m.ShouldDownload(tc.chapter())
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}

			if tc.logInclude != "" {
				log := buffer.String()
				if !strings.Contains(log, tc.logInclude) {
					t.Errorf("Failed to log: got %s, want %s", log, tc.logInclude)
				}
			}

			if tc.after != nil {
				tc.after(t, m)
			}

		})
	}

}

func TestChapterSearchResponse_FilterOneEnChapter(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{
		GetChaptersFunc: func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
			return &ChapterSearchResponse{
				Data: []ChapterSearchData{
					{
						Attributes: ChapterAttributes{
							Chapter:            "4",
							TranslatedLanguage: "en",
						},
					},
					{
						Attributes: ChapterAttributes{
							Chapter:            "4",
							TranslatedLanguage: "fr",
						},
					},
				},
			}, nil
		},
	})
	m.language = "en"

	res, err := m.repository.GetChapters(t.Context(), RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	filtered := m.FilterChapters(res)
	if len(filtered.Data) != 1 {
		t.Errorf("Expected 1 chapters, got %d", len(filtered.Data))
	}
}

func TestChapterSearchResponse_FilterOneEnChapterSkipOfficial(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})
	m.language = "en"

	c := ChapterSearchResponse{
		Result:   "",
		Response: "",
		Data: []ChapterSearchData{
			{
				Attributes: ChapterAttributes{
					Chapter:            "4",
					ExternalUrl:        "some external url",
					TranslatedLanguage: "en",
				},
			},
		},
		Limit:  0,
		Offset: 0,
		Total:  0,
	}

	filtered := m.FilterChapters(&c)
	if len(filtered.Data) != 0 {
		t.Errorf("Expected 0 chapters, got %d", len(filtered.Data))
	}

}

func TestTagsBlackList(t *testing.T) {
	m := tempManga(t, req(), io.Discard, &mockRepository{})
	m.language = "en"

	chpt := chapter()
	m.Preference = &models.Preference{
		BlackListedTags: []models.Tag{
			{
				Name:           "Blacklisted Genre",
				NormalizedName: "blacklistedgenre",
			},
			{
				Name:           "Blacklisted Tag",
				NormalizedName: "blacklistedtag",
			},
			{
				Name:           "ABC",
				NormalizedName: "abc",
			},
		},
	}

	m.info = mangaResp()

	ci := m.comicInfo(chpt)
	genres := strings.Split(ci.Genre, ",")
	tags := strings.Split(ci.Tags, ",")

	got := len(genres)
	want := 1

	if want != got {
		t.Errorf("want %d genres, got %d: %+v", want, got, genres)
	}

	got = len(tags)
	want = 1
	if want != got {
		t.Errorf("want %d tags, got %d: %+v", want, got, tags)
	}

}

func TestReplaceCover(t *testing.T) {
	m := tempManga(t, payload.DownloadRequest{
		Provider:  models.MANGADEX,
		Id:        "9b8e6611-73cd-4a50-883b-79ba99b7e4b3",
		BaseDir:   "Manga",
		TempTitle: "Finder Goshi no Ano Ko",
		DownloadMetadata: models.DownloadRequestMetadata{
			StartImmediately: false,
		},
		IsSubscription: false,
	}, io.Discard, &mockRepository{
		GetCoverImagesFunc: func(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error) {
			return &MangaCoverResponse{
				Data: []MangaCoverData{
					{
						Attributes: MangaCoverAttributes{
							Volume:   "2",
							FileName: "d34b33fd-91c6-4a8a-b015-1c0cfc580ad6.jpg",
						},
					},
				},
			}, nil
		},
		GetChaptersFunc: func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
			return &ChapterSearchResponse{
				Data: []ChapterSearchData{
					{
						Id: "f811f708-6b2b-45cb-b639-ee3fd186fc39",
						Attributes: ChapterAttributes{
							Chapter:            "7",
							Volume:             "2",
							TranslatedLanguage: "en",
						},
					},
				},
			}, nil
		},
		GetChapterImagesFunc: func(ctx context.Context, id string) (*ChapterImageSearchResponse, error) {
			// They have a system with random servers
			return tempRepo(t, io.Discard).GetChapterImages(ctx, id)
		}})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Fatal("m.LoadInfo() timeout")
	}

	chapterSeven := utils.Find(m.chapters.Data, func(data ChapterSearchData) bool {
		return data.Attributes.Chapter == "7"
	})

	if chapterSeven == nil {
		t.Fatal("chapterSeven is nil")
	}

	originalCover, ok := m.coverFactory(chapterSeven.Attributes.Volume)
	if !ok {
		t.Fatal("chapterSeven.Attributes.Volume cover not available")
	}

	coverBytes, _, err := m.getBetterChapterCover(*chapterSeven, originalCover)
	if err != nil {
		t.Fatal(err)
	}

	if len(originalCover.Bytes) != len(coverBytes) {
		t.Fatal("Cover should not have been replaced, but was")
	}

}

func TestReplaceCoverDestructionLover(t *testing.T) {
	t.Skipf("Removed from mangadex")
	m := tempManga(t, payload.DownloadRequest{
		Provider:  models.MANGADEX,
		Id:        "192f8421-55f9-46e2-9c6e-87a88c2f048a",
		BaseDir:   "Manga",
		TempTitle: "Destruction Lover",
		DownloadMetadata: models.DownloadRequestMetadata{
			StartImmediately: false,
		},
		IsSubscription: false,
	}, io.Discard, &mockRepository{
		GetCoverImagesFunc: func(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error) {
			return &MangaCoverResponse{
				Data: []MangaCoverData{
					{
						Attributes: MangaCoverAttributes{
							Volume:   "2",
							FileName: "a9f46f7a-fb0a-448a-8d21-b0434ade2037.jpg",
						},
					},
				},
			}, nil
		},
		GetChaptersFunc: func(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
			return &ChapterSearchResponse{
				Data: []ChapterSearchData{
					{
						Id: "9d7673d5-caee-4c19-8844-31af2133de96",
						Attributes: ChapterAttributes{
							Chapter:            "5",
							Volume:             "2",
							TranslatedLanguage: "en",
						},
					},
				},
			}, nil
		},
		GetChapterImagesFunc: func(ctx context.Context, id string) (*ChapterImageSearchResponse, error) {
			// They have a system with random servers
			return tempRepo(t, io.Discard).GetChapterImages(ctx, id)
		}})

	select {
	case <-m.LoadInfo(t.Context()):
		break
	case <-time.After(waitTimeOut):
		t.Fatal("m.LoadInfo() timeout")
	}

	chapterFive := utils.Find(m.chapters.Data, func(data ChapterSearchData) bool {
		return data.Attributes.Chapter == "5"
	})

	if chapterFive == nil {
		t.Fatal("chapterFive is nil")
	}

	originalCover, ok := m.coverFactory(chapterFive.Attributes.Volume)
	if !ok {
		t.Fatalf("chapterSeven.Attributes.Volume (%s) cover not available", chapterFive.Attributes.Volume)
	}

	coverBytes, firstPage, err := m.getBetterChapterCover(*chapterFive, originalCover)
	if err != nil {
		t.Fatal(err)
	}

	if len(originalCover.Bytes) == len(coverBytes) {
		t.Fatal("Cover should have been replaced, but wasn't")
	}

	if !firstPage {
		t.Fatal("Expected firstPage to be a valid cover")
	}

}
