package services

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"strconv"
	"time"
)

type SettingsService interface {
	GetSettingsDto() (payload.Settings, error)
	UpdateSettingsDto(settings payload.Settings) error
}

type settingsService struct {
	settings models.Settings
	log      zerolog.Logger

	cachedSettings utils.CachedItem[payload.Settings]
}

func SettingsServiceProvider(settings models.Settings, log zerolog.Logger) SettingsService {
	return &settingsService{
		settings: settings,
		log:      log.With().Str("handler", "settings-service").Logger(),
	}
}

func (s *settingsService) GetSettingsDto() (payload.Settings, error) {
	if !s.cachedSettings.HasExpired() {
		return s.cachedSettings.Get()
	}

	settings, err := s.settings.All()
	if err != nil {
		return payload.Settings{}, err
	}

	var dto payload.Settings
	for _, setting := range settings {
		if err = s.ParseSetting(setting, &dto); err != nil {
			return payload.Settings{}, err
		}
	}

	s.cachedSettings = utils.NewCachedItem(dto, time.Minute*10)

	return dto, nil
}

func (s *settingsService) UpdateSettingsDto(dto payload.Settings) error {
	settings, err := s.settings.All()
	if err != nil {
		return err
	}

	for _, setting := range settings {
		if err = s.SerializeSetting(&setting, dto); err != nil {
			return err
		}
	}

	if err = s.settings.Update(settings); err != nil {
		return err
	}
	s.cachedSettings.SetExpired()
	return nil
}

func (s *settingsService) SerializeSetting(setting *models.ServerSetting, dto payload.Settings) error {
	var err error
	switch setting.Key {
	case models.RootDir:
		setting.Value = dto.RootDir
	case models.BaseUrl:
		setting.Value = dto.BaseUrl
	case models.RedisAddr:
		setting.Value = dto.RedisAddr
	case models.CacheType:
		setting.Value = string(dto.CacheType)
	case models.DisableIpv6:
		setting.Value = strconv.FormatBool(dto.DisableIpv6)
	case models.MaxConcurrentImages:
		setting.Value = strconv.Itoa(dto.MaxConcurrentImages)
	case models.MaxConcurrentTorrents:
		setting.Value = strconv.Itoa(dto.MaxConcurrentTorrents)
	case models.OidcAuthority:
		setting.Value = dto.Oidc.Authority
	case models.OidcClientID:
		setting.Value = dto.Oidc.ClientID
	}

	return err
}

func (s *settingsService) ParseSetting(setting models.ServerSetting, dto *payload.Settings) error {
	var err error
	switch setting.Key {
	case models.RootDir:
		dto.RootDir = setting.Value
	case models.BaseUrl:
		dto.BaseUrl = setting.Value
	case models.RedisAddr:
		dto.RedisAddr = setting.Value
	case models.CacheType:
		dto.CacheType = config.CacheType(setting.Value)
	case models.DisableIpv6:
		dto.DisableIpv6, err = strconv.ParseBool(setting.Value)
	case models.MaxConcurrentImages:
		dto.MaxConcurrentImages, err = strconv.Atoi(setting.Value)
	case models.MaxConcurrentTorrents:
		dto.MaxConcurrentTorrents, err = strconv.Atoi(setting.Value)
	case models.OidcAuthority:
		dto.Oidc.Authority = setting.Value
	case models.OidcClientID:
		dto.Oidc.ClientID = setting.Value
	}

	return err
}
