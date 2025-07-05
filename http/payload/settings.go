package payload

import "github.com/Fesaa/Media-Provider/config"

type Settings struct {
	BaseUrl               string           `json:"baseUrl"`
	CacheType             config.CacheType `json:"cacheType"`
	RedisAddr             string           `json:"redisAddr"`
	MaxConcurrentTorrents int              `json:"maxConcurrentTorrents"`
	MaxConcurrentImages   int              `json:"maxConcurrentImages"`
	DisableIpv6           bool             `json:"disableIpv6"`
	RootDir               string           `json:"rootDir"`
	Oidc                  OidcSettings     `json:"oidc"`
}

type OidcSettings struct {
	Authority            string `json:"authority"`
	ClientID             string `json:"clientId"`
	DisablePasswordLogin bool   `json:"disablePasswordLogin"`
	AutoLogin            bool   `json:"autoLogin"`
}
