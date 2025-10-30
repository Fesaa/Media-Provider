package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SubscriptionService interface {
	// Add a new subscription, saved to DB and starts the cron job
	// Subscription is normalized in the process
	Add(context.Context, models.Subscription) (*models.Subscription, error)
	// Update an existing subscription, updates DB. Subscription is normalized in the process
	Update(context.Context, models.Subscription) error
	// Delete the subscription with ID
	Delete(context.Context, int) error

	// UpdateHour recreates the underlying cronjob. Generally only called when the hour to run subscriptions changes
	UpdateHour(ctx context.Context) error
}

type subscriptionService struct {
	cronService    CronService
	contentService ContentService
	notifier       NotificationService
	settings       SettingsService
	transloco      TranslocoService

	unitOfWork *db.UnitOfWork
	log        zerolog.Logger

	job gocron.Job
}

func SubscriptionServiceProvider(unitOfWork *db.UnitOfWork, provider ContentService,
	log zerolog.Logger, cronService CronService, notifier NotificationService,
	transloco TranslocoService, ctx context.Context, settings SettingsService,
) (SubscriptionService, error) {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanSetupService,
		trace.WithAttributes(attribute.String("service.name", "SubscriptionService")))
	defer span.End()

	service := &subscriptionService{
		cronService:    cronService,
		contentService: provider,
		settings:       settings,
		notifier:       notifier,
		transloco:      transloco,
		unitOfWork:     unitOfWork,
		log:            log.With().Str("handler", "subscription-service").Logger(),
	}

	if err := service.OnStartUp(ctx); err != nil {
		return nil, fmt.Errorf("SubscriptionService OnStartUp: %w", err)
	}

	return service, nil
}

func (s *subscriptionService) UpdateHour(ctx context.Context) error {
	return s.OnStartUp(ctx)
}

func (s *subscriptionService) OnStartUp(ctx context.Context) error {
	subs, err := s.unitOfWork.Subscriptions.All(ctx)
	if err != nil {
		return err
	}

	settings, err := s.settings.GetSettingsDto(ctx)
	if err != nil {
		return err
	}

	err = s.unitOfWork.Transaction(func(unitOfWork *db.UnitOfWork) error {
		for _, sub := range subs {
			sub.NextExecution = sub.GetNextExecution(settings.SubscriptionRefreshHour)
			if err = unitOfWork.Subscriptions.Update(ctx, sub); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return s.UpdateTask(ctx, settings.SubscriptionRefreshHour)
}

func (s *subscriptionService) orFromPreferences(ctx context.Context, hours ...int) (int, error) {
	if len(hours) > 0 {
		return hours[0], nil
	}

	settings, err := s.settings.GetSettingsDto(ctx)
	if err != nil {
		return 0, err
	}

	return settings.SubscriptionRefreshHour, nil
}

func (s *subscriptionService) UpdateTask(ctx context.Context, hours ...int) error {
	hour, err := s.orFromPreferences(ctx, hours...)
	if err != nil {
		return err
	}

	if s.job != nil {
		if err = s.cronService.RemoveJob(s.job.ID()); err != nil {
			return err
		}
	}

	cronString := fmt.Sprintf("0 %d * * *", hour)
	s.log.Debug().Str("cronString", cronString).Msg("scheduling subscription job with cron string")

	job, err := s.cronService.NewJob(gocron.CronJob(cronString, false), s.subscriptionTask(hour))
	if err != nil {
		return err
	}

	s.job = job
	return nil
}

func (s *subscriptionService) Add(ctx context.Context, sub models.Subscription) (*models.Subscription, error) {
	existing, err := s.unitOfWork.Subscriptions.GetByContentID(ctx, sub.ContentId)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("subscription already exists")
	}

	settings, err := s.settings.GetSettingsDto(ctx)
	if err != nil {
		return nil, err
	}

	sub.Normalize(settings.SubscriptionRefreshHour)
	sub.LastCheck = time.Now()
	sub.LastCheckSuccess = true
	sub.NextExecution = sub.GetNextExecution(settings.SubscriptionRefreshHour)

	newSub, err := s.unitOfWork.Subscriptions.New(ctx, sub)
	if err != nil {
		return nil, err
	}

	return newSub, nil
}

func (s *subscriptionService) Update(ctx context.Context, sub models.Subscription) error {
	cur, err := s.unitOfWork.Subscriptions.GetByContentID(ctx, sub.ContentId)
	if err != nil {
		return err
	}

	if cur == nil {
		return errors.New("subscription doesn't exist")
	}

	settings, err := s.settings.GetSettingsDto(ctx)
	if err != nil {
		return err
	}

	// Reset no download count when refresh frequency changes
	if cur.RefreshFrequency != sub.RefreshFrequency {
		cur.NoDownloadCount = 0
	}

	cur.Title = sub.Title
	cur.BaseDir = sub.BaseDir
	cur.RefreshFrequency = sub.RefreshFrequency
	cur.Provider = sub.Provider
	cur.Payload = sub.Payload
	cur.LastDownloadDir = sub.LastDownloadDir

	cur.Normalize(settings.SubscriptionRefreshHour)
	cur.NextExecution = sub.GetNextExecution(settings.SubscriptionRefreshHour)
	s.log.Debug().Time("nextExecution", sub.NextExecution).
		Msg("subscription will run next on")

	return s.unitOfWork.Subscriptions.Update(ctx, *cur)
}

func (s *subscriptionService) Delete(ctx context.Context, id int) error {
	return s.unitOfWork.Subscriptions.Delete(ctx, id)
}

func (s *subscriptionService) subscriptionTask(hour int) gocron.Task {
	s.log.Debug().Int("hour", hour).Msg("creating subscription task")
	return gocron.NewTask(func(ctx context.Context) {
		ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesSubscriptionTask)
		defer span.End()

		s.log.Debug().Msg("running subscription task")

		subs, err := s.unitOfWork.Subscriptions.All(ctx)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to get subscriptions")
			s.notifier.Notify(ctx, models.NewNotification().
				WithTitle(s.transloco.GetTranslation("failed-to-run-subscriptions")).
				WithBody(s.transloco.GetTranslation("failed-to-run-subscriptions-body", err)).
				WithGroup(models.GroupError).
				WithColour(models.Error).
				WithRequiredRoles(models.ManageSubscriptions).
				Build())
			return
		}

		counter := 0
		now := time.Now()
		for _, sub := range subs {
			nextExec := sub.NextExecution.In(time.Local)
			if !utils.IsSameDay(now, nextExec) {
				s.log.Debug().Time("nextExec", nextExec).
					Time("now", now).Msg("next execution is on a different date. Skipping")
				// Subscription only run once a day, if these don't match. It's for another day.
				continue
			}

			s.handleSub(ctx, sub, hour)
			counter++
		}

		s.log.Debug().Int("counter", counter).Msg("ran subscriptions")
	})
}

func (s *subscriptionService) handleSub(ctx context.Context, sub models.Subscription, hour int) {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesSubscriptionTask+".run",
		trace.WithAttributes(attribute.Int("id", sub.ID)))
	defer span.End()

	err := s.contentService.DownloadSubscription(&sub)
	sub.LastCheck = time.Now()
	sub.LastCheckSuccess = err == nil
	sub.NextExecution = sub.GetNextExecution(hour)

	if err != nil {
		s.log.Error().Err(err).
			Int("id", sub.ID).
			Str("contentId", sub.ContentId).
			Msg("failed to download content")
		s.notifier.Notify(ctx, models.NewNotification().
			WithTitle(s.transloco.GetTranslation("failed-sub")).
			WithBody(s.transloco.GetTranslation("failed-start-sub-download", sub.Title, err)).
			WithGroup(models.GroupError).
			WithColour(models.Error).
			WithRequiredRoles(models.ManageSubscriptions).
			Build())
		return
	}

	if err = s.unitOfWork.Subscriptions.Update(ctx, sub); err != nil {
		s.log.Warn().Err(err).Int("id", sub.ID).Msg("failed to update subscription")
	} else {
		s.log.Debug().Int("id", sub.ID).Msg("updated subscription")
	}
}
