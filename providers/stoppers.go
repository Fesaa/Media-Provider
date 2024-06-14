package providers

import (
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/yoitsu"
)

func yoitsuStopper(req StopRequest) error {
	return yoitsu.I().RemoveDownload(req.Id, req.DeleteFiles)
}

func mangadexStopper(req StopRequest) error {
	return mangadex.I().RemoveDownload(req.Id, req.DeleteFiles)
}
