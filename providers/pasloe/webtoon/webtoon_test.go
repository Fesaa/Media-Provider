package webtoon

import (
	"bytes"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
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

func req() payload.DownloadRequest {
	return payload.DownloadRequest{
		Provider:  models.WEBTOON,
		Id:        utils.Stringify(WebToonID),
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

func tempWebtoon(t *testing.T, w io.Writer, dirs ...string) *webtoon {
	t.Helper()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	mock := mockClient{baseDir: utils.OrDefault(dirs, t.TempDir())}

	cont := dig.New()
	scope := cont.Scope("tempWebtoon")

	must(scope.Provide(func() api.Client {
		return mock
	}))
	must(scope.Provide(utils.Identity(http.DefaultClient)))
	must(scope.Provide(utils.Identity(zerolog.New(w))))
	must(scope.Provide(utils.Identity(req())))
	must(scope.Provide(NewRepository))

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

func (m *mockRepository) Search(options SearchOptions) ([]SearchData, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.searchRes, nil
}

func (m *mockRepository) LoadImages(chapter Chapter) ([]string, error) {
	if m.imageErr != nil {
		return nil, m.imageErr
	}
	return m.imageRes, nil
}

func (m *mockRepository) SeriesInfo(id string) (*Series, error) {
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
	case <-wt.LoadInfo():
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
	case <-wt.LoadInfo():
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
	case <-wt.LoadInfo():
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
			Id:   0,
			Name: "An other series",
		},
	}

	select {
	case <-wt.LoadInfo():
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
	case <-wt.LoadInfo():
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

	want := WebToonName + " Ch. 8"
	got := wt.ContentDir(chapter())
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWebtoon_ContentPath(t *testing.T) {
	wt := tempWebtoon(t, io.Discard)

	want := path.Join(WebToonName, WebToonName+" Ch. 8")
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

func TestWebtoon_ContentLogger(t *testing.T) {
	var buffer bytes.Buffer
	wt := tempWebtoon(t, &buffer)

	l := wt.ContentLogger(chapter())
	l.Info().Msg("a")

	want := "{\"level\":\"info\",\"handler\":\"webtoon\",\"id\":\"4747\",\"number\":\"8\",\"title\":\"Episode 8\",\"message\":\"a\"}\n"
	got := buffer.String()
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
	got, err := wt.ContentUrls(chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != want {
		t.Errorf("got %d, want %d", len(got), want)
	}

	mock.imageErr = fmt.Errorf("error")
	got, err = wt.ContentUrls(chapter())
	if err == nil {
		t.Errorf("got %v, want error", got)
	}

}

func TestWebtoon_WriteContentMetaData(t *testing.T) {
	var buf bytes.Buffer
	dir := t.TempDir()
	w := tempWebtoon(t, &buf, dir)
	w.info = series()

	if err := os.MkdirAll(w.ContentPath(chapter()), 0755); err != nil {
		t.Fatal(err)
	}

	if err := w.WriteContentMetaData(chapter()); err != nil {
		t.Fatal(err)
	}

	ciPath := path.Join(w.ContentPath(chapter()), "ComicInfo.xml")
	_, err := os.Stat(ciPath)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatal(err)
	}

	ci := w.comicInfo()
	var b bytes.Buffer
	if err = comicinfo.Write(ci, &b); err != nil {
		t.Fatal(err)
	}

	if b.String() != string(data) {
		t.Errorf("m.comicInfo() = %q, want %q", b, data)
	}

	buf.Reset()
}

func TestWebtoon_DownloadContent(t *testing.T) {
	var buffer bytes.Buffer
	w := tempWebtoon(t, &buffer)

	urls, err := w.ContentUrls(chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) == 0 {
		t.Fatal("len(urls) = 0, want > 0")
	}

	if err = os.MkdirAll(path.Join(w.ContentPath(chapter())), 0755); err != nil {
		t.Fatal(err)
	}

	if err = w.DownloadContent(1, chapter(), urls[0]); err != nil {
		t.Fatal(err)
	}

	filePath := path.Join(w.ContentPath(chapter()), fmt.Sprintf("page %s.jpg", utils.PadInt(1, 4)))
	_, err = os.Stat(filePath)
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
			if wt.ContentRegex().MatchString(tt.s) != tt.match {
				t.Errorf("got %v, want %v", wt.ContentRegex().MatchString(tt.s), tt.match)
			}
		})
	}
}
