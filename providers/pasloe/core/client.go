package core

import (
	"github.com/Fesaa/Media-Provider/db/models"
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
	GetConfig() Config
	CanStart(models.Provider) bool
}
