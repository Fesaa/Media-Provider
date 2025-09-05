package payload

import "github.com/Fesaa/Media-Provider/config"

type Settings struct {
	BaseUrl               string           `json:"baseUrl"`
	CacheType             config.CacheType `json:"cacheType" validate:"required,oneof=MEMORY REDIS"`
	RedisAddr             string           `json:"redisAddr"`
	MaxConcurrentTorrents int              `json:"maxConcurrentTorrents" validate:"required,min=1,max=10"`
	MaxConcurrentImages   int              `json:"maxConcurrentImages" validate:"required,min=1,max=5"`
	DisableIpv6           bool             `json:"disableIpv6"`
	RootDir               string           `json:"rootDir"`
	Oidc                  OidcSettings     `json:"oidc"`
}

type OidcSettings struct {
	Authority            string `json:"authority"`
	ClientID             string `json:"clientId"`
	ClientSecret         string `json:"clientSecret"`
	RedirectURL          string `json:"redirectUrl"`
	DisablePasswordLogin bool   `json:"disablePasswordLogin"`
	AutoLogin            bool   `json:"autoLogin"`
}

func (o OidcSettings) Enabled() bool {
	return o.Authority != "" && o.ClientID != "" && o.ClientSecret != "" && o.RedirectURL != ""
}

type PublicOidcSettings struct {
	DisablePasswordLogin bool `json:"disablePasswordLogin"`
	AutoLogin            bool `json:"autoLogin"`
	Enabled              bool `json:"enabled"`
}
