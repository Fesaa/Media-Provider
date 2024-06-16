package providers

import (
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/yoitsu"
)

func yoitsuDownloader(req payload.DownloadRequest) error {
	_, err := yoitsu.I().AddDownload(req)
	return err
}

func mangadexDownloader(req payload.DownloadRequest) error {
	return mangadex.I().Download(req)
}
