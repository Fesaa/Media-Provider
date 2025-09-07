package mock

import (
	"context"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
)

type Settings struct {
	settings *payload.Settings
}

func (s *Settings) UpdateCurrentVersion(ctx context.Context) error {
	return nil
}

func (s *Settings) GetSettingsDto(context.Context) (payload.Settings, error) {
	if s.settings == nil {
		return payload.Settings{
			BaseUrl:               "",
			CacheType:             config.MEMORY,
			RedisAddr:             "",
			MaxConcurrentTorrents: 5,
			MaxConcurrentImages:   5,
			DisableIpv6:           false,
			RootDir:               "temp",
			Oidc: payload.OidcSettings{
				Authority:            "",
				ClientID:             "",
				DisablePasswordLogin: false,
				AutoLogin:            false,
			},
		}, nil
	}
	return *s.settings, nil
}

func (s *Settings) UpdateSettingsDto(ctx context.Context, settings payload.Settings) error {
	s.settings = &settings
	return nil
}
