package models

import "github.com/anacrolix/torrent"

type TorrentProvider interface {
	GetBackingClient() *torrent.Client

	AddDownload(infoHash string) (*torrent.Torrent, error)
	RemoveDownload(infoHash string) error
	GetRunningTorrents() map[string]*torrent.Torrent
}
