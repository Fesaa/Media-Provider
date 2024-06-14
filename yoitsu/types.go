package yoitsu

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
)

// TorrentInfo contains information about a torrent
// This is the information that is sent to the client
type TorrentInfo struct {
	Provider  config.Provider `json:"Provider"`
	InfoHash  string          `json:"InfoHash"`
	Name      string          `json:"Name"`
	Size      int64           `json:"Size"`
	Progress  int64           `json:"Progress"`
	Completed int64           `json:"Completed"`
	Speed     string          `json:"Speed"`
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
	GetInfo() TorrentInfo
}

// Yoitsu wrapper around the torrent.Client struct
type Yoitsu interface {
	// GetBackingClient returns the torrent.Client
	GetBackingClient() *torrent.Client

	// AddDownload adds a new download to the client.
	// The baseDir is the directory where the files will be downloaded to.
	// Should not include specific location, just the base directory. The torrent hash will be appended to it.
	// Returns the torrent that was added
	AddDownload(infoHash string, baseDir string, provider config.Provider) (Torrent, error)

	// RemoveDownload removes a download from the wrapper, optionally deleting the files
	RemoveDownload(infoHash string, deleteFiles bool) error

	// GetRunningTorrents returns a map of all running torrents, indexed by their info hash
	GetRunningTorrents() *utils.SafeMap[string, Torrent]

	GetBaseDir() string
}
