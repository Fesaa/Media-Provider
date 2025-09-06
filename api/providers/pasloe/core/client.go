package core

import (
	"github.com/Fesaa/Media-Provider/services"
)

type Config interface {
	GetRootDir() string
	GetMaxConcurrentImages() int
}

type Client interface {
	services.Client
	GetBaseDir() string
	GetCurrentDownloads() []Downloadable
	MoveToDownloadQueue(id string) error

	Shutdown() error
}
