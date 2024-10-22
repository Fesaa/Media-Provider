package providers

import (
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/providers/mangadex"
	"github.com/Fesaa/Media-Provider/providers/webtoon"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
)

func yoitsuDownloader(req payload.DownloadRequest) error {
	_, err := yoitsu.I().AddDownload(req)
	return err
}

func mangadexDownloader(req payload.DownloadRequest) error {
	_, err := mangadex.I().Download(req)
	return err
}

func webToonDownloader(req payload.DownloadRequest) error {
	_, err := webtoon.I().Download(req)
	return err
}
