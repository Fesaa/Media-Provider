package core

import (
	"context"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"io"
	"testing"
)

type PasloeClient struct {
	BaseDir string
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
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) Provider() models.Provider {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) LoadInfo(ctx context.Context) chan struct{} {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) GetInfo() payload.InfoStat {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) All() []ChapterMock {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) ContentList() []payload.ListContentData {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) ContentDir(t ChapterMock) string {
	return p.contentDir
}

func (p ProviderMock) ContentUrls(ctx context.Context, t ChapterMock) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) WriteContentMetaData(t ChapterMock) error {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) IsContent(s string) bool {
	//TODO implement me
	panic("implement me")
}

func (p ProviderMock) ShouldDownload(t ChapterMock) bool {
	//TODO implement me
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

	tempDir := utils.OrElse(req.BaseDir, t.TempDir())
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
	must(scope.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))
	must(scope.Provide(services.ArchiveServiceProvider))
	return New[ChapterMock, *SeriesMock](scope, "test", provider)
}
