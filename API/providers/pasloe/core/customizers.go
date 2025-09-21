package core

import (
	"context"
	"net/http"
)

type DownloadCustomizer interface {
	// CustomizeRequest allows providers to add custom headers or modify the request
	CustomizeRequest(req *http.Request) error
}

type ChapterCustomizer[C Chapter] interface {
	// CustomizeAllChapters allows a provider to set a custom method to retrieve all chapter, in case Series does not have the data
	CustomizeAllChapters() []C
}

type PreDownloadHook interface {
	CustomizePreDownloadHook(context.Context)
}
