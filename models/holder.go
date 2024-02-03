package models

const HolderKey string = "holder"

type Holder interface {
	GetAuthProvider() AuthProvider
	GetDatabaseProvider() DatabaseProvider
	GetTorrentProvider() TorrentProvider

	Shutdown() error
}
