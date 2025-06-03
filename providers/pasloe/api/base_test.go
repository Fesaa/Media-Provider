package api

import (
	"github.com/Fesaa/Media-Provider/db/models"
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

func testBase(t *testing.T, req payload.DownloadRequest, w io.Writer) *DownloadBase[IDAble] {
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
	client := PasloeClient{BaseDir: tempDir}

	must(scope.Provide(utils.Identity(req)))
	must(scope.Provide(func() Client {
		return &client
	}))
	must(scope.Provide(utils.Identity(log)))
	must(scope.Provide(func() services.SignalRService { return nil }))
	must(scope.Provide(func() services.NotificationService { return nil }))
	must(scope.Provide(func() services.TranslocoService { return nil }))
	must(scope.Provide(func() models.Preferences { return nil }))
	must(scope.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))

	return NewBaseWithProvider[IDAble](scope, "test", nil)
}
