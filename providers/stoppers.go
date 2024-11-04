package providers

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
)

func yoitsuStopper(req payload.StopRequest) error {
	return yoitsu.I().RemoveDownload(req)
}

func mangadexStopper(req payload.StopRequest) error {
	return pasloe.I().RemoveDownload(req)
}

func webtoonStopper(req payload.StopRequest) error {
	return pasloe.I().RemoveDownload(req)
}
