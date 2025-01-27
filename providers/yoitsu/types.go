package yoitsu

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
)

type Config interface {
	GetRootDir() string
	GetMaxConcurrentTorrents() int
}

// Torrent wrapper around torrent.Torrent
type Torrent interface {
	// GetTorrent returns the wrapped torrent.Torrent
	GetTorrent() *torrent.Torrent
	// WaitForInfoAndDownload loads the torrent.Torrent's info and starts downloading afters.
	//
	// You may cancel this with Torrent.Cancel
	WaitForInfoAndDownload()
	// Cancel stops WaitForInfoAndDownload, returns an error if it wasn't started yet
	Cancel() error
	// GetInfo returns useful information about the torrent
	GetInfo() payload.InfoStat
	GetDownloadDir() string
}

// Yoitsu wrapper around the torrent.Client struct
type Yoitsu interface {
	// GetBackingClient returns the torrent.Client
	GetBackingClient() *torrent.Client

	// AddDownload adds a new download to the client.
	// A nill error does NOT mean Torrent is not nill, if the torrent is queued
	// both will be nill
	AddDownload(payload.DownloadRequest) (Torrent, error)

	// RemoveDownload removes a download from the wrapper, optionally deleting the files
	RemoveDownload(request payload.StopRequest) error

	// GetRunningTorrents returns a map of all running torrents, indexed by their info hash
	GetRunningTorrents() *utils.SafeMap[string, Torrent]
	GetQueuedTorrents() []payload.InfoStat

	GetBaseDir() string
}
