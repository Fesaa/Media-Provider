package models

type SettingKey int

const (
	OidcAuthority SettingKey = iota
	OidcClientID
	BaseUrl
	CacheType
	RedisAddr
	MaxConcurrentTorrents
	MaxConcurrentImages
	DisableIpv6
	RootDir
)

type ServerSetting struct {
	Key   SettingKey `gorm:"primary_key"`
	Value string     `gorm:"type:text"`
}
