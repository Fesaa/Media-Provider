package models

const HolderKey string = "holder"

type Holder interface {
	GetAuthProvider() AuthProvider
	GetStorageProvider() StorageProvider
	GetDatabaseProvider() DatabaseProvider

	Shutdown() error
}
