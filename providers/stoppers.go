package providers

import (
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/webtoon"
	"github.com/Fesaa/Media-Provider/yoitsu"
)

func yoitsuStopper(req payload.StopRequest) error {
	return yoitsu.I().RemoveDownload(req)
}

func mangadexStopper(req payload.StopRequest) error {
	return mangadex.I().RemoveDownload(req)
}

func webtoonStopper(req payload.StopRequest) error {
	return webtoon.I().RemoveDownload(req)
}
