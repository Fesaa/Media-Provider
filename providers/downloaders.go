package providers

import "github.com/Fesaa/Media-Provider/yoitsu"

func yoitsuDownloader(req DownloadRequest) error {
	_, err := yoitsu.I().AddDownload(req.Hash, req.BaseDir)
	return err
}
