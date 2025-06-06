package core

import (
	"net/http"
	"regexp"
)

type DownloadCustomizer interface {
	// CustomizeRequest allows providers to add custom headers or modify the request
	CustomizeRequest(req *http.Request) error
}

type ContentCustomizer interface {
	// GetVolumeSubDir returns the volume subdirectory path (empty if not used)
	GetVolumeSubDir(chapter Chapter) string
	// GetContentRegexPatterns returns regex patterns for content validation
	GetContentRegexPatterns() []*regexp.Regexp
}
