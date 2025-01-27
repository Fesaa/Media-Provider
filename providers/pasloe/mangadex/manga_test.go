package mangadex

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

type mockClient struct {
	baseDir string
}

func (m mockClient) GetRootDir() string {
	return m.baseDir
}

func (m mockClient) GetMaxConcurrentImages() int {
	return 5
}

func (m mockClient) Download(request payload.DownloadRequest) error {
	return nil
}

func (m mockClient) RemoveDownload(request payload.StopRequest) error {
	return nil
}

func (m mockClient) GetBaseDir() string {
	return m.baseDir
}

func (m mockClient) GetCurrentDownloads() []api.Downloadable {
	return []api.Downloadable{}
}

func (m mockClient) GetQueuedDownloads() []payload.QueueStat {
	return []payload.QueueStat{}
}

func (m mockClient) GetConfig() api.Config {
	return m
}

type mockRepo struct {
	t           *testing.T
	manga       GetMangaResponse
	chapters    ChapterSearchResponse
	images      ChapterImageSearchResponse
	covers      MangaCoverResponse
	mangaErr    error
	chaptersErr error
	imagesErr   error
	coversErr   error
}

func (m mockRepo) GetManga(id string) (*GetMangaResponse, error) {
	return &m.manga, m.mangaErr
}

func (m mockRepo) SearchManga(options SearchOptions) (*MangaSearchResponse, error) {
	return &MangaSearchResponse{}, nil
}

func (m mockRepo) GetChapters(id string, offset ...int) (*ChapterSearchResponse, error) {
	return &m.chapters, m.chaptersErr
}

func (m mockRepo) GetChapterImages(id string) (*ChapterImageSearchResponse, error) {
	return &m.images, m.imagesErr
}

func (m mockRepo) GetCoverImages(id string, offset ...int) (*MangaCoverResponse, error) {
	return &m.covers, m.coversErr
}

func tempManga(t *testing.T, req payload.DownloadRequest, w io.Writer, td ...string) *manga {
	t.Helper()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	log := zerolog.New(w)

	c := dig.New()
	scope := c.Scope("testScope")

	tempDir := utils.OrDefault(td, t.TempDir())
	client := mockClient{baseDir: tempDir}
	repo := tempRepo(t, w)

	must(scope.Provide(func() api.Client {
		return &client
	}))
	must(scope.Provide(utils.Identity(log)))
	must(scope.Provide(utils.Identity(http.DefaultClient)))
	must(scope.Provide(utils.Identity(repo)))
	must(scope.Provide(utils.Identity(req)))
	must(scope.Provide(services.MarkdownServiceProvider))

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
			},
		},
	}
}

func TestManga_Title(t *testing.T) {
	m := tempManga(t, req(), io.Discard)

	want := RainbowsAfterStormsID
	got := m.Title()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
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
	m := tempManga(t, req(), &buf)

	m.id = "DFGHJKJHGFDGHJKHGFGHJ"
	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
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
	m := tempManga(t, req(), &buf)

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for manga title")
	}

	mock := mockRepo{t: t}
	mock.manga = GetMangaResponse{
		Data: MangaSearchData{
			Attributes: MangaAttributes{
				LastChapter: "8",
				LastVolume:  "2",
			},
		},
	}
	mock.chapters = ChapterSearchResponse{
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
	}

	m.repository = mock

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for manga title")
	}

	if !m.foundLastVolume {
		t.Error("expected manga to have last chapter")
	}

	if !m.foundLastChapter {
		t.Error("expected manga to have last chapter")
	}

	got := m.totalVolumes
	want := 2

	if got != want {
		t.Errorf("totalVolumes got %d, want %d", got, want)
	}

	got = m.totalChapters
	want = 2
	if got != want {
		t.Errorf("totalChapters got %d, want %d", got, want)
	}

}

//nolint:funlen
func TestManga_LoadInfoErrors(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf)

	mock := mockRepo{t: t}
	mock.manga = GetMangaResponse{Data: MangaSearchData{Attributes: MangaAttributes{}}}
	mock.mangaErr = errors.New("error")
	mock.chapters = ChapterSearchResponse{}
	mock.chaptersErr = errors.New("error")

	m.repository = mock

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for manga title")
	}

	log := buf.String()
	if !strings.Contains(log, "error while loading manga info") {
		t.Errorf("got %q, want 'error while loading manga info'", log)
	}
	buf.Reset()

	mock.mangaErr = nil
	m.repository = mock
	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for manga title")
	}

	log = buf.String()
	if !strings.Contains(log, "error while loading chapter info") {
		t.Errorf("got %q, want 'error while loading chapter info'", log)
	}
	buf.Reset()

	mock.chaptersErr = nil
	mock.coversErr = errors.New("error")
	m.repository = mock

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for manga title")
	}
	log = buf.String()
	if !strings.Contains(log, "error while loading manga coverFactory") {
		t.Errorf("got %q, want 'error while loading manga coverFactory'", log)
	}
	buf.Reset()

	mock.chapters = ChapterSearchResponse{
		Data: []ChapterSearchData{
			{Attributes: ChapterAttributes{
				TranslatedLanguage: "en",
				Volume:             "NotANumber",
			}},
		},
	}

	m.repository = mock
	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for manga title")
	}

	log = buf.String()
	if !strings.Contains(log, "not adding chapter, as Volume string isn't an int") {
		t.Errorf("got %q, want 'not adding chapter, as Volume isn't an int'", log)
	}

}

func TestManga_Provider(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	if m.Provider() != models.MANGADEX {
		t.Errorf("got %q, want %q", m.Provider(), models.MANGADEX)
	}
}

func TestManga_All(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	mock := mockRepo{t: t}
	mock.chapters = ChapterSearchResponse{
		Data: []ChapterSearchData{
			{Attributes: ChapterAttributes{TranslatedLanguage: "en"}},
		},
	}
	m.repository = mock

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for manga title")
	}

	got := m.All()
	want := 1

	if len(got) != want {
		t.Errorf("got %d, want %d", len(got), want)
	}

}

func TestManga_ContentDir(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	m.info = mangaResp()

	got := m.ContentDir(chapter())
	want := "Rainbows After Storms Ch. 0162"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestManga_ContentDirBadChapter(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf)
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
	want = "Rainbows After Storms OneShot"
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
	m := tempManga(t, req(), io.Discard)
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
	m := tempManga(t, req(), io.Discard)

	got := m.ContentKey(chapter())
	want := RainbowsAfterStormsLastChapterID
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestManga_ContentLogger(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf)

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
	m := tempManga(t, req(), io.Discard)

	mock := mockRepo{t: t}
	mock.images = ChapterImageSearchResponse{
		Chapter: ChapterInfo{
			Data: []string{"1", "2", "3"},
		},
	}

	m.repository = mock

	urls, err := m.ContentUrls(chapter())
	if err != nil {
		t.Fatal(err)
	}

	want := 3
	if len(urls) != want {
		t.Errorf("got %d, want %d", len(urls), want)
	}

	mock.imagesErr = errors.New("error")
	m.repository = mock
	_, err = m.ContentUrls(chapter())
	if err == nil {
		t.Errorf("got %v, want error", urls)
	}

}

func TestManga_WriteContentMetaData(t *testing.T) {
	var buf bytes.Buffer
	dir := t.TempDir()
	m := tempManga(t, req(), &buf, dir)
	m.info = mangaResp()
	m.coverFactory = func(volume string) (string, bool) {
		return "", false
	}

	if err := m.WriteContentMetaData(chapter()); err != nil {
		t.Fatal(err)
	}

	ciPath := path.Join(m.ContentPath(chapter()), "comicinfo.xml")
	_, err := os.Stat(ciPath)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatal(err)
	}

	ci := m.comicInfo(chapter())
	var b bytes.Buffer
	if err = comicinfo.Write(ci, &b); err != nil {
		t.Fatal(err)
	}

	if b.String() != string(data) {
		t.Errorf("m.comicInfo() = %q, want %q", b, data)
	}

	buf.Reset()
	if err = m.WriteContentMetaData(chapter()); err != nil {
		t.Fatal(err)
	}

	log := buf.String()
	want := "volume metadata already written, skipping"
	if !strings.Contains(log, want) {
		t.Errorf("got %s, want %s", log, want)
	}
}

func TestManga_writeCIStatusNotCompleted(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	ci := comicinfo.NewComicInfo()

	m.info = mangaResp()
	m.info.Attributes.Status = StatusCancelled

	m.writeCIStatus(ci)
	if ci.Count != 0 {
		t.Errorf("got %d, want 0", ci.Count)
	}

}

func TestManga_writeCIStatusOnlyChapters(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	ci := comicinfo.NewComicInfo()

	m.info = mangaResp()
	m.info.Attributes.Status = StatusCompleted
	m.totalVolumes = 0
	m.totalChapters = 10
	m.foundLastChapter = true

	m.writeCIStatus(ci)
	if ci.Count != 10 {
		t.Errorf("got %d, want 10", ci.Count)
	}

}

func TestManga_writeCIStatusVolumes(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	ci := comicinfo.NewComicInfo()

	m.info = mangaResp()
	m.info.Attributes.Status = StatusCompleted
	m.foundLastChapter = true
	m.foundLastVolume = true
	m.totalVolumes = 12

	m.writeCIStatus(ci)
	if ci.Count != 12 {
		t.Errorf("got %d, want 12", ci.Count)
	}

}

func TestManga_writeCIStatusDontWarnTwice(t *testing.T) {
	var buf bytes.Buffer
	m := tempManga(t, req(), &buf)
	ci := comicinfo.NewComicInfo()

	m.info = mangaResp()
	m.info.Attributes.Status = StatusCompleted

	m.writeCIStatus(ci)
	m.writeCIStatus(ci)
	m.writeCIStatus(ci)

	count := strings.Count(buf.String(), "Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?")
	if count != 1 {
		t.Errorf("got %d, want 1", count)
	}

}

func TestManga_DownloadContent(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

	urls, err := m.ContentUrls(chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) == 0 {
		t.Fatal("len(urls) = 0, want > 0")
	}

	if err = os.MkdirAll(path.Join(m.ContentPath(chapter())), 0755); err != nil {
		t.Fatal(err)
	}

	if err = m.DownloadContent(1, chapter(), urls[0]); err != nil {
		t.Fatal(err)
	}

	filePath := path.Join(m.ContentPath(chapter()), fmt.Sprintf("page %s.jpg", utils.PadInt(1, 4)))
	_, err = os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestManga_ContentRegex(t *testing.T) {
	m := tempManga(t, req(), io.Discard)

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
	ios := services.IOServiceProvider(zerolog.Nop())

	type test struct {
		name          string
		contentOnDisk []api.Content
		chapter       func() ChapterSearchData
		command       func(*testing.T, *manga)
		want          bool
		logInclude    string
		after         func(*testing.T, *manga)
	}

	tests := []test{
		{
			name:          "New Download",
			contentOnDisk: []api.Content{},
			chapter:       chapter,
			want:          true,
		},
		{
			name: "Volume on disk",
			contentOnDisk: []api.Content{
				{Name: RainbowsAfterStorms + " Vol. 13.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13.cbz")},
			},
			chapter: chapter,
			want:    false,
		},
		{
			name: "Chapter on disk, no volume",
			contentOnDisk: []api.Content{
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
			contentOnDisk: []api.Content{
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
			contentOnDisk: []api.Content{
				{Name: RainbowsAfterStorms + " Ch. 0162.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13", RainbowsAfterStorms+" Ch. 0162.cbz")},
			},
			chapter: chapter,
			command: func(t *testing.T, m *manga) {
				t.Helper()
				fullpath := path.Join(m.Client.GetBaseDir(), RainbowsAfterStorms, RainbowsAfterStorms+" Vol. 13", RainbowsAfterStorms+" Ch. 0162")
				ci := comicinfo.NewComicInfo()
				ci.Volume = 13
				if err := os.MkdirAll(fullpath, 0755); err != nil {
					t.Fatal(err)
				}
				if err := comicinfo.Save(ci, path.Join(fullpath, "comicinfo.xml")); err != nil {
					t.Fatal(err)
				}

				if err := ios.ZipFolder(fullpath, fullpath+".cbz"); err != nil {
					t.Fatal(err)
				}
			},
			want:       false,
			logInclude: "Volume on disk matches, not replacing",
			after:      nil,
		},
		{
			name: "Chapter on disk, replacing loose chapter",
			contentOnDisk: []api.Content{
				{Name: RainbowsAfterStorms + " Ch. 0162.cbz", Path: path.Join(RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 0162.cbz")},
			},
			chapter: chapter,
			command: func(t *testing.T, m *manga) {
				t.Helper()
				fullpath := path.Join(m.Client.GetBaseDir(), RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 0162")
				ci := comicinfo.NewComicInfo()
				if err := os.MkdirAll(fullpath, 0755); err != nil {
					t.Fatal(err)
				}
				if err := comicinfo.Save(ci, path.Join(fullpath, "comicinfo.xml")); err != nil {
					t.Fatal(err)
				}

				if err := ios.ZipFolder(fullpath, fullpath+".cbz"); err != nil {
					t.Fatal(err)
				}
			},
			want:       true,
			logInclude: "Loose chapter has been assigned to a volume, replacing",
			after: func(t *testing.T, m *manga) {
				t.Helper()
				fullpath := path.Join(m.Client.GetBaseDir(), RainbowsAfterStorms, RainbowsAfterStorms+" Ch. 162.cbz")
				_, err := os.Stat(fullpath)
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("expected %s to not exist", fullpath)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			m := tempManga(t, req(), &buffer)
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
	m := tempManga(t, req(), io.Discard)
	m.language = "en"

	res, err := m.repository.GetChapters(RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	filtered := m.FilterChapters(res)
	if len(filtered.Data) != 172 {
		t.Errorf("Expected 172 chapters, got %d", len(filtered.Data))
	}
}

func TestChapterSearchResponse_FilterOneEnChapterSkipOfficial(t *testing.T) {
	m := tempManga(t, req(), io.Discard)
	m.language = "en"

	c := ChapterSearchResponse{
		Result:   "",
		Response: "",
		Data: []ChapterSearchData{
			{
				Attributes: ChapterAttributes{
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
