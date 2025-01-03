package api

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
	"regexp"
)

type Downloadable interface {
	Title() string
	Id() string
	Provider() models.Provider
	GetBaseDir() string
	Cancel()
	GetInfo() payload.InfoStat
	WaitForInfoAndDownload()
	GetDownloadDir() string
	GetOnDiskContent() []string
}

type DownloadInfoProvider[T any] interface {
	Title() string
	Provider() models.Provider
	LoadInfo() chan struct{}
	GetInfo() payload.InfoStat
	All() []T

	ContentDir(t T) string
	ContentPath(t T) string
	ContentKey(t T) string
	ContentLogger(t T) zerolog.Logger

	ContentUrls(t T) ([]string, error)
	WriteContentMetaData(t T) error
	DownloadContent(idx int, t T, url string) error

	ContentRegex() *regexp.Regexp
	ShouldDownload(t T) bool
}
