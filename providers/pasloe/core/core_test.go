package core

import (
	"context"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"io"
	"reflect"
	"testing"
)

type PasloeClient struct {
	BaseDir string
}

type SettingsService struct {
	settings *payload.Settings
}

func (s *SettingsService) GetSettingsDto() (payload.Settings, error) {
	if s.settings == nil {
		return payload.Settings{
			BaseUrl:               "",
			CacheType:             config.MEMORY,
			RedisAddr:             "",
			MaxConcurrentTorrents: 5,
			MaxConcurrentImages:   5,
			DisableIpv6:           false,
			RootDir:               "temp",
			Oidc: payload.OidcSettings{
				Authority:            "",
				ClientID:             "",
				DisablePasswordLogin: false,
				AutoLogin:            false,
			},
		}, nil
	}
	return *s.settings, nil
}

func (s *SettingsService) UpdateSettingsDto(settings payload.Settings) error {
	s.settings = &settings
	return nil
}

func (m PasloeClient) GetRootDir() string {
	return m.BaseDir
}

func (m PasloeClient) GetMaxConcurrentImages() int {
	return 5
}

func (m PasloeClient) Download(request payload.DownloadRequest) error {
	return nil
}

func (m PasloeClient) RemoveDownload(request payload.StopRequest) error {
	return nil
}

func (m PasloeClient) GetBaseDir() string {
	return m.BaseDir
}

func (m PasloeClient) GetCurrentDownloads() []Downloadable {
	return []Downloadable{}
}

func (m PasloeClient) GetQueuedDownloads() []payload.InfoStat {
	return []payload.InfoStat{}
}

func (m PasloeClient) GetConfig() Config {
	return m
}

func (m PasloeClient) Content(id string) services.Content {
	return nil
}

func (m PasloeClient) CanStart(models.Provider) bool {
	return true
}

func req() payload.DownloadRequest {
	return payload.DownloadRequest{
		Provider:  models.DYNASTY,
		Id:        "spice_and_wolf",
		BaseDir:   "",
		TempTitle: "Spice and Wolf",
	}
}

type ProviderMock struct {
	title      string
	contentDir string
}

func (p ProviderMock) Title() string {
	return p.title
}

func (p ProviderMock) RefUrl() string {
	panic("implement me")
}

func (p ProviderMock) Provider() models.Provider {
	panic("implement me")
}

func (p ProviderMock) LoadInfo(ctx context.Context) chan struct{} {
	panic("implement me")
}

func (p ProviderMock) ContentUrls(ctx context.Context, t ChapterMock) ([]string, error) {
	panic("implement me")
}

func (p ProviderMock) WriteContentMetaData(ctx context.Context, t ChapterMock) error {
	panic("implement me")
}

func testBase(t *testing.T, req payload.DownloadRequest, w io.Writer, provider ProviderMock) *Core[ChapterMock, *SeriesMock] {
	t.Helper()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	log := zerolog.New(w)

	c := dig.New()
	scope := c.Scope("testScope")

	tempDir := utils.OrElse(req.BaseDir, "")
	req.BaseDir = "" // Reset base dir so prevent stacking
	client := PasloeClient{BaseDir: tempDir}

	must(scope.Provide(utils.Identity(req)))
	must(scope.Provide(utils.Identity(menou.DefaultClient)))
	must(scope.Provide(func() Client {
		return &client
	}))
	must(scope.Provide(utils.Identity(log)))
	must(scope.Provide(func() services.SignalRService { return nil }))
	must(scope.Provide(func() services.NotificationService { return nil }))
	must(scope.Provide(func() services.TranslocoService { return nil }))
	must(scope.Provide(func() models.Preferences { return nil }))
	must(scope.Provide(func() services.SettingsService { return &SettingsService{} }))
	must(scope.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))
	must(scope.Provide(services.ArchiveServiceProvider))
	must(scope.Provide(services.ImageServiceProvider))
	return New[ChapterMock, *SeriesMock](scope, "test", provider)
}

func TestContentList_EmptyChapters(t *testing.T) {
	core := &Core[ChapterMock, SeriesMock]{
		SeriesInfo: SeriesMock{
			chapters: []ChapterMock{},
		},
	}

	got := core.ContentList()
	if len(got) != 0 {
		t.Errorf("Expected nil, got %#v", got)
	}
}

func TestContentList_SingleChapterNoVolume(t *testing.T) {
	chpt := ChapterMock{Id: "1", Title: "MockSeries", LabelStr: "Ch 1", Volume: "", Chapter: "1"}
	core := &Core[ChapterMock, SeriesMock]{
		SeriesInfo: SeriesMock{
			chapters: []ChapterMock{chpt},
		},
		ToDownload: []ChapterMock{chpt},
		impl:       ProviderMock{title: "MockSeries"},
	}

	got := core.ContentList()
	if len(got) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(got))
	}
	expectedLabel := "MockSeries Ch 1"
	if got[0].Label != expectedLabel {
		t.Errorf("Expected label %q, got %q", expectedLabel, got[0].Label)
	}
	if got[0].SubContentId != "1" {
		t.Errorf("Expected SubContentId '1', got %q", got[0].SubContentId)
	}
	if !got[0].Selected {
		t.Error("Expected Selected true, got false")
	}
}

func TestContentList_SingleChapterWithVolume(t *testing.T) {
	core := &Core[ChapterMock, SeriesMock]{
		SeriesInfo: SeriesMock{
			chapters: []ChapterMock{
				{Id: "1", Title: "", LabelStr: "Ch 1", Volume: "1", Chapter: "1"},
			},
		},
		impl: ProviderMock{title: "MockSeries"},
	}

	got := core.ContentList()
	if len(got) != 1 {
		t.Fatalf("Expected 1 volume group, got %d", len(got))
	}
	if got[0].Label != "Volume 1" {
		t.Errorf("Expected label 'Volume 1', got %q", got[0].Label)
	}
	if len(got[0].Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(got[0].Children))
	}
	if got[0].Children[0].Label != "Ch 1" {
		t.Errorf("Expected child label 'Ch 1', got %q", got[0].Children[0].Label)
	}
}

func TestContentList_MultipleVolumesSorted(t *testing.T) {
	core := &Core[ChapterMock, SeriesMock]{
		SeriesInfo: SeriesMock{
			chapters: []ChapterMock{
				{Id: "1", Title: "T", LabelStr: "C1", Volume: "2", Chapter: "1"},
				{Id: "2", Title: "T", LabelStr: "C2", Volume: "1", Chapter: "1"},
			},
		},
		impl: ProviderMock{title: "MockSeries"},
	}

	got := core.ContentList()
	if len(got) != 2 {
		t.Fatalf("Expected 2 volume groups, got %d", len(got))
	}
	if got[0].Label != "Volume 2" || got[1].Label != "Volume 1" {
		t.Errorf("Expected volume labels 'Volume 1' and 'Volume 2', got %q and %q", got[0].Label, got[1].Label)
	}
}

func TestContentList_SelectedFiltering(t *testing.T) {
	core := &Core[ChapterMock, SeriesMock]{
		SeriesInfo: SeriesMock{
			chapters: []ChapterMock{
				{Id: "1", Title: "", LabelStr: "C1", Volume: "", Chapter: "1"},
				{Id: "2", Title: "", LabelStr: "C2", Volume: "", Chapter: "2"},
			},
		},
		ToDownloadUserSelected: []string{"2"},
		impl:                   ProviderMock{title: "MockSeries"},
	}

	got := core.ContentList()
	if len(got) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(got))
	}

	gotSelected := []bool{got[0].Selected, got[1].Selected}
	expected := []bool{true, false}
	if !reflect.DeepEqual(gotSelected, expected) {
		t.Errorf("Expected selected flags %v, got %v", expected, gotSelected)
	}
}
