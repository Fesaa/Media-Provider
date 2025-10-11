package publication

import (
	"context"
	"net/http"
)

type PreDownloadHook interface {
	PreDownloadHook(Publication, context.Context) error
}

type HttpGetHook interface {
	HttpGetHook(*http.Request) error
}
