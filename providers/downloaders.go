package providers

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
)

func yoitsuDownloader(req payload.DownloadRequest) error {
	_, err := yoitsu.I().AddDownload(req)
	return err
}

func mangadexDownloader(req payload.DownloadRequest) error {
	_, err := pasloe.I().Download(req)
	return err
}

func webToonDownloader(req payload.DownloadRequest) error {
	_, err := pasloe.I().Download(req)
	return err
}
