package core

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
)

type DisplayInformation struct {
	Name string
}

type Downloadable interface {
	services.Content

	Logger() *zerolog.Logger

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
	// CleanupNewContent takes a path from GetNewContent to clean up
	CleanupNewContent(string) error

	// LoadMetadata loads all required metadata to start the download , this method blocks until complete or cancelled
	LoadMetadata(ctx context.Context)
	// DownloadContent starts the download process, this method blocks until complete or cancelled
	DownloadContent(ctx context.Context)

	// GetNewContentNamed returns the names of the downloaded content (chapters)
	GetNewContentNamed() []string

	FailedDownloads() int
}

type DownloadInfoProvider[C Chapter, S Series[C]] interface {
	Provider() models.Provider
	Title() string
	RefUrl() string
	LoadInfo(context.Context) chan struct{}

	ContentUrls(context.Context, C) ([]string, error)
	WriteContentMetaData(context.Context, C) error
	CoreExt() Ext[C, S]
}

type IOTaskFunc[C Chapter, S Series[C]] func(*Core[C, S], context.Context, zerolog.Logger, IOTask) error

type CleanupFunc[C Chapter, S Series[C]] func(*Core[C, S], string) error

type IsContentFunc func(string) (Content, bool)

type VolumeFunc[C Chapter, S Series[C]] func(*Core[C, S], Content) (string, error)

type Ext[C Chapter, S Series[C]] struct {
	IoTaskFunc         IOTaskFunc[C, S]
	ContentCleanupFunc CleanupFunc[C, S]
	IsContentFunc      IsContentFunc
	VolumeFunc         VolumeFunc[C, S]
}
