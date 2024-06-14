package providers

import (
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/yoitsu"
)

func yoitsuDownloader(req DownloadRequest) error {
	_, err := yoitsu.I().AddDownload(req.Hash, req.BaseDir, req.Provider)
	return err
}

func mangadexDownloader(req DownloadRequest) error {
	return mangadex.I().Download(req.Hash, req.BaseDir)
}
