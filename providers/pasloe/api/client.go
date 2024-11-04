package api

import "github.com/Fesaa/Media-Provider/http/payload"

type Config interface {
	GetRootDir() string
	GetMaxConcurrentImages() int
}

type Client interface {
	Download(request payload.DownloadRequest) (Downloadable, error)
	RemoveDownload(request payload.StopRequest) error

	GetBaseDir() string
	GetCurrentDownloads() []Downloadable
	GetQueuedDownloads() []payload.QueueStat
	GetConfig() Config
}
