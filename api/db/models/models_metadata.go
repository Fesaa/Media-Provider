package models

type MetadataRow struct {
	Key   MetadataKey
	Value string
}

type MetadataKey int

const (
	InstalledVersion MetadataKey = iota
	FirstInstalledVersion
	InstallDate
)
