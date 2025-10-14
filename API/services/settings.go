package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/internal/metadata"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type SettingsService interface {
	GetSettingsDto(context.Context) (payload.Settings, error)
	// UpdateSettingsDto updates the current settings
	UpdateSettingsDto(context.Context, payload.Settings) error
	UpdateCurrentVersion(ctx context.Context) error
}

type settingsService struct {
	unitOfWork          *db.UnitOfWork
	subscriptionService SubscriptionService
	signalR             SignalRService
	transLoco           TranslocoService
	log                 zerolog.Logger

	cachedSettings utils.CachedItem[payload.Settings]
}

func SettingsServiceProvider(
	unitOfWork *db.UnitOfWork,
	subService SubscriptionService,
	signalR SignalRService,
	transLoco TranslocoService,
	log zerolog.Logger,
) SettingsService {
	return &settingsService{
		unitOfWork:          unitOfWork,
		subscriptionService: subService,
		signalR:             signalR,
		transLoco:           transLoco,
		log:                 log.With().Str("handler", "settings-service").Logger(),
	}
}

func (s *settingsService) UpdateCurrentVersion(ctx context.Context) error {
	return s.unitOfWork.Transaction(func(unitOfWork *db.UnitOfWork) error {
		err := unitOfWork.Settings.Update(ctx, []models.ServerSetting{{
			Key:   models.InstalledVersion,
			Value: metadata.Version.String(),
		}})
		if err != nil {
			return err
		}

		return unitOfWork.Settings.Update(ctx, []models.ServerSetting{{
			Key:   models.LastUpdateDate,
			Value: time.Now().Format(time.RFC3339),
		}})
	})
}

func (s *settingsService) GetSettingsDto(ctx context.Context) (payload.Settings, error) {
	if s.cachedSettings != nil && !s.cachedSettings.HasExpired() {
		s.log.Trace().Msg("settings being returned from cache")
		return s.cachedSettings.Get()
	}

	settings, err := s.unitOfWork.Settings.GetAll(ctx)
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

func (s *settingsService) UpdateSettingsDto(ctx context.Context, dto payload.Settings) error {
	cur, err := s.GetSettingsDto(ctx)
	if err != nil {
		return err
	}

	settings, err := s.unitOfWork.Settings.GetAll(ctx)
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

	if err = s.unitOfWork.Settings.Update(ctx, settings); err != nil {
		return err
	}

	if cur.SubscriptionRefreshHour != dto.SubscriptionRefreshHour {
		if err = s.subscriptionService.UpdateTask(ctx, dto.SubscriptionRefreshHour); err != nil {
			s.log.Error().Err(err).Msg("failed to update subscription refresh hour")
			s.signalR.Notify(ctx, models.NewNotification().
				WithTitle(s.transLoco.GetTranslation("failed-to-register-sub-task-title")).
				WithSummary(s.transLoco.GetTranslation("failed-to-register-sub-task-summary")).
				WithOwner(contextkey.GetFromCtxOrDefault(ctx, contextkey.User).ID).
				WithGroup(models.GroupError).
				WithColour(models.Error).
				Build())
		}
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
		if dto.Oidc.ClientSecret != strings.Repeat("*", len(setting.Value)) {
			setting.Value = dto.Oidc.ClientSecret
		}
	case models.SubscriptionRefreshHour:
		setting.Value = strconv.Itoa(dto.SubscriptionRefreshHour)
	case models.InstalledVersion:
	case models.FirstInstalledVersion:
	case models.InstallDate:
	case models.DbDriver:
	case models.LastUpdateDate:
		break // Do not update
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
	case models.InstalledVersion:
		dto.Metadata.Version = metadata.SemanticVersion(setting.Value)
	case models.FirstInstalledVersion:
		dto.Metadata.FirstInstalledVersion = setting.Value
	case models.InstallDate:
		dto.Metadata.InstallDate, err = time.Parse(time.RFC3339, setting.Value)
	case models.SubscriptionRefreshHour:
		dto.SubscriptionRefreshHour, err = strconv.Atoi(setting.Value)
	case models.LastUpdateDate:
		dto.Metadata.LastUpdateDate, err = time.Parse(time.RFC3339, setting.Value)
	case models.DbDriver:
		break // ignore
	}

	return err
}
