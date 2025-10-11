package publication

import (
	"context"

	"github.com/Fesaa/Media-Provider/http/payload"
)

type Repository interface {
	// SeriesInfo returns all series information for the given id
	SeriesInfo(context.Context, string, payload.DownloadRequest) (Series, error)
	// ChapterUrls returns all urls that need to be downloaded to complete the chapters
	ChapterUrls(context.Context, Chapter) ([]string, error)
}
