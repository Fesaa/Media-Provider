package core

import (
	"context"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
)

type DisplayInformation struct {
	Name string
}

type Downloadable interface {
	services.Content
	GetBaseDir() string
	Cancel()
	GetDownloadDir() string

	DisplayInformation() DisplayInformation

	// GetOnDiskContent returns the name of the files that have been identified as already existing content
	GetOnDiskContent() []Content
	// GetNewContent returns the full (relative) path of downloaded content.
	// This will be a slice of paths produced by DownloadInfoProvider.ContentPath
	GetNewContent() []string
	// GetToRemoveContent returns the full (relative) path of old content that has to be removed
	GetToRemoveContent() []string

	StartLoadInfo()
	StartDownload()

	// GetNewContentNamed returns the names of the downloaded content (chapters)
	GetNewContentNamed() []string

	FailedDownloads() int
}

type DownloadInfoProvider[T any] interface {
	Title() string
	RefUrl() string
	Provider() models.Provider
	LoadInfo(ctx context.Context) chan struct{}
	GetInfo() payload.InfoStat
	All() []T
	ContentList() []payload.ListContentData

	ContentDir(t T) string
	ContentPath(t T) string
	ContentKey(t T) string
	ContentLogger(t T) zerolog.Logger

	ContentUrls(ctx context.Context, t T) ([]string, error)
	WriteContentMetaData(t T) error
	DownloadContent(idx int, t T, url string) error

	IsContent(string) bool
	ShouldDownload(t T) bool
}
