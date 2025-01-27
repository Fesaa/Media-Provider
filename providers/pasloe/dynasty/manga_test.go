package dynasty

import (
	"bytes"
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

const (
	SailorGirlFriend = "Sailor Girlfriend"

	// For tests with special non-chapter; https://dynasty-scans.com/series/shiawase_trimming
	ShiawaseTrimming   = "Shiawase Trimming"
	ShiawaseTrimmingId = "shiawase_trimming"
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

func (m mockClient) GetQueuedDownloads() []payload.InfoStat {
	return []payload.InfoStat{}
}

func (m mockClient) GetConfig() api.Config {
	return m
}

func tempManga(t *testing.T, req payload.DownloadRequest, w io.Writer) *manga {
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
	client := mockClient{baseDir: tempDir}
	repo := tempRepository(w)

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

	m := tempManga(t, req(), &buffer)

	want := "sailor_girlfriend"
	if m.Title() != want {
		t.Errorf("m.Title() = %q, want %q", m.Title(), want)
	}

	select {
	case <-m.LoadInfo():
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
	m := tempManga(t, r, &buffer)

	select {
	case <-m.LoadInfo():
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
	m := tempManga(t, req(), &buffer)

	want := models.DYNASTY
	if m.Provider() != want {
		t.Errorf("m.Provider() = %q, want %q", m.Provider(), want)
	}
}

func TestManga_GetInfoBeforeLoad(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

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
	m := tempManga(t, req(), &buffer)

	select {
	case <-m.LoadInfo():
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
	m := tempManga(t, req(), &buffer)

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	got := m.All()
	if len(got) != 5 {
		t.Errorf("len(got) = %d, want %d", len(got), 5)
	}
}

func TestManga_ContentDir(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

	got := m.ContentDir(chapter())
	want := "sailor_girlfriend Ch. 0004.5"

	if got != want {
		t.Errorf("m.ContentDir() = %q, want %q", got, want)
	}

}

func TestManga_ContentPath(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

	got := m.ContentPath(chapter())
	want := path.Join(m.Client.GetBaseDir(), "sailor_girlfriend/sailor_girlfriend Ch. 0004.5")
	if got != want {
		t.Errorf("m.ContentPath() = %q, want %q", got, want)
	}
}

func TestManga_ContentKey(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

	got := m.ContentKey(chapter())
	want := "sailor_girlfriend_ch4_5"
	if got != want {
		t.Errorf("m.ContentKey() = %q, want %q", got, want)
	}
}

func TestManga_ContentLogger(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

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
	m := tempManga(t, req(), &buffer)

	urls, err := m.ContentUrls(chapter())
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != 19 {
		t.Errorf("len(urls) = %d, want %d", len(urls), 19)
	}

}

func TestManga_WriteContentMetaData(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

	if err := os.MkdirAll(path.Join(m.ContentPath(chapter())), 0755); err != nil {
		t.Fatal(err)
	}

	err := m.WriteContentMetaData(chapter())
	if err != nil {
		t.Fatal(err)
	}

	p := path.Join(m.ContentPath(chapter()), "ComicInfo.xml")
	_, err = os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(p)
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
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)

	if m.IsContent("Not a Valid Chapter") {
		t.Error("m.ContentRegex().MatchString() returned true")
	}

	if !m.IsContent("Sailor Girlfriend Ch. 0004.5.cbz") {
		t.Error("m.ContentRegex().MatchString() returned false")
	}

	if !m.IsContent("Shiawase Trimming OneShot Manga Time Kirara 20th Anniversary Special Collaboration: Stardust Telepath x Shiawase Trimming.cbz") {
		t.Error("m.ContentRegex().MatchString() returned false")
	}
}

func TestManga_ShouldDownload(t *testing.T) {
	var buffer bytes.Buffer
	m := tempManga(t, req(), &buffer)
	m.ExistingContent = []api.Content{
		{
			Name: "Sailor Girlfriend Ch. 0001.cbz",
		},
		{
			Name: "Sailor Girlfriend Ch. 0002.cbz",
		},
	}

	select {
	case <-m.LoadInfo():
		break
	case <-time.After(5 * time.Second):
		t.Fatal("m.LoadInfo() timeout")
	}

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
