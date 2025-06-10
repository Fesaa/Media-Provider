package core

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
)

type Series[C Chapter] interface {
	GetId() string
	GetTitle() string
	AllChapters() []C
}

func (c *Core[C, S]) GetAllLoadedChapters() []C {
	if chapterCustomizer, ok := c.impl.(ChapterCustomizer[C]); ok {
		return chapterCustomizer.CustomizeAllChapters()
	}

	return c.SeriesInfo.AllChapters()
}

func (c *Core[C, S]) Title() string {
	// Need to offload the interface and logic to a separate function to prevent compile errors
	// as we cannot compare S with nil, for some reason.
	return computeTitle(c.SeriesInfo, c.Req)
}

// computeTitle returns the title of the series
func computeTitle(series interface{ GetTitle() string }, req payload.DownloadRequest) string {
	if req.IsSubscription && req.Sub.Info.Title != "" {
		return req.Sub.Info.Title
	}

	if series == nil {
		return utils.NonEmpty(req.TempTitle, req.Id)
	}

	return utils.NonEmpty(series.GetTitle(), req.TempTitle, req.Id)
}
