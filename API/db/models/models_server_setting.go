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
	OidcAutoLogin
	OidcDisablePasswordLogin
	OidcClientSecret
)

type ServerSetting struct {
	Key   SettingKey `gorm:"unique"`
	Value string     `gorm:"type:text"`
}
