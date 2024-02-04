package models

const HolderKey string = "holder"

type Holder interface {
	GetAuthProvider() AuthProvider
	GetTorrentProvider() TorrentProvider

	Shutdown() error
}
