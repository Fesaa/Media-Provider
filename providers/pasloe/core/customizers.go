package core

import (
	"net/http"
)

type DownloadCustomizer interface {
	// CustomizeRequest allows providers to add custom headers or modify the request
	CustomizeRequest(req *http.Request) error
}
