package services

import (
	"errors"
	"strconv"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
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
	if s.cachedSettings != nil && !s.cachedSettings.HasExpired() {
		s.log.Trace().Msg("settings being returned from cache")
		return s.cachedSettings.Get()
	}

	settings, err := s.settings.All()
	if err != nil {
		return payload.Settings{}, err
	}

	var dto payload.Settings
	for _, setting := range settings {
		if err = s.parseSetting(setting, &dto); err != nil {
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

	var errs []error
	settings = utils.Map(settings, func(t models.ServerSetting) models.ServerSetting {
		if err = s.serializeSetting(&t, dto); err != nil {
			errs = append(errs, err)
		}
		return t
	})
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if err = s.settings.Update(settings); err != nil {
		return err
	}
	s.cachedSettings.SetExpired()
	return nil
}

func (s *settingsService) serializeSetting(setting *models.ServerSetting, dto payload.Settings) error {
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
	case models.OidcAutoLogin:
		setting.Value = strconv.FormatBool(dto.Oidc.AutoLogin)
	case models.OidcDisablePasswordLogin:
		setting.Value = strconv.FormatBool(dto.Oidc.DisablePasswordLogin)
	case models.OidcClientSecret:
		setting.Value = dto.Oidc.ClientSecret
	case models.OidcRedirectUrl:
		setting.Value = dto.Oidc.RedirectURL
	}

	return err
}

func (s *settingsService) parseSetting(setting models.ServerSetting, dto *payload.Settings) error {
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
	case models.OidcAutoLogin:
		dto.Oidc.AutoLogin, err = strconv.ParseBool(setting.Value)
	case models.OidcDisablePasswordLogin:
		dto.Oidc.DisablePasswordLogin, err = strconv.ParseBool(setting.Value)
	case models.OidcClientSecret:
		dto.Oidc.ClientSecret = setting.Value
	case models.OidcRedirectUrl:
		dto.Oidc.RedirectURL = setting.Value
	}

	return err
}
