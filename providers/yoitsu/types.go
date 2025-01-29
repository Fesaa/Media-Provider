package yoitsu

import (
	"github.com/Fesaa/Media-Provider/http/payload"
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
	// GetTorrent returns the wrapped torrent.Torrent
	GetTorrent() *torrent.Torrent
	// WaitForInfoAndDownload loads the torrent.Torrent's info and starts downloading afters.
	//
	// You may cancel this with Torrent.Cancel
	WaitForInfoAndDownload()
	// Cancel stops WaitForInfoAndDownload, returns an error if it wasn't started yet
	Cancel() error
	GetDownloadDir() string
}

// Yoitsu wrapper around the torrent.Client struct
type Yoitsu interface {
	services.Client
	// GetBackingClient returns the torrent.Client
	GetBackingClient() *torrent.Client

	// GetRunningTorrents returns a map of all running torrents, indexed by their info hash
	GetRunningTorrents() utils.SafeMap[string, Torrent]
	GetQueuedTorrents() []payload.InfoStat

	GetBaseDir() string
}
