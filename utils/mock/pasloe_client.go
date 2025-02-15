package mock

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
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

func (m PasloeClient) GetCurrentDownloads() []api.Downloadable {
	return []api.Downloadable{}
}

func (m PasloeClient) GetQueuedDownloads() []payload.InfoStat {
	return []payload.InfoStat{}
}

func (m PasloeClient) GetConfig() api.Config {
	return m
}

func (m PasloeClient) Content(id string) services.Content {
	return nil
}

func (m PasloeClient) CanStart(models.Provider) bool {
	return true
}
