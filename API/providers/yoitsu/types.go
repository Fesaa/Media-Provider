package yoitsu

import (
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
)

type Config interface {
	GetRootDir() string
	GetMaxConcurrentTorrents() int
}

// Torrent wrapper around torrent.Torrent
type Torrent interface {
	services.Content
	GetTorrent() *torrent.Torrent
	LoadInfo()
	StartDownload()
	Cancel()
	IsDone() bool
	Cleanup(root string)
	Files() int
}

// Client wrapper around the torrent.Client struct
type Client interface {
	services.Client
	GetTorrents() utils.SafeMap[string, Torrent]
	CanStartNext() bool
	Shutdown() error
}
