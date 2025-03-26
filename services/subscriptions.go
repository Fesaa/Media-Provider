package services

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"time"
)

type SubscriptionService interface {
	// Get the subscription with ID
	Get(uint) (*models.Subscription, error)
	// All returns all active subscriptions
	All() ([]models.Subscription, error)
	// Add a new subscription, saved to DB and starts the cron job
	// Subscription is normalized in the process
	Add(models.Subscription) (*models.Subscription, error)
	// Update an existing subscription, updates DB. Subscription is normalized in the process
	Update(models.Subscription) error
	// Delete the subscription with ID
	Delete(uint) error

	// UpdateTask recreates the underlying cronjob. Generally only called when the hour to run susbcriptions changes
	UpdateTask(hour ...int) error
}

type subscriptionService struct {
	cronService    CronService
	contentService ContentService
	notifier       NotificationService
	transloco      TranslocoService

	db  *db.Database
	log zerolog.Logger

	job gocron.Job
}

func SubscriptionServiceProvider(db *db.Database, provider ContentService,
	log zerolog.Logger, cronService CronService, notifier NotificationService,
	transloco TranslocoService,
) (SubscriptionService, error) {
	service := &subscriptionService{
		cronService:    cronService,
		contentService: provider,
		notifier:       notifier,
		transloco:      transloco,
		db:             db,
		log:            log.With().Str("handler", "subscription-service").Logger(),
	}

	if err := service.UpdateTask(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *subscriptionService) orFromPreferences(hours ...int) (int, error) {
	if len(hours) > 0 {
		return hours[0], nil
	}

	pref, err := s.db.Preferences.Get()
	if err != nil {
		return 0, err
	}

	return pref.SubscriptionRefreshHour, nil
}

func (s *subscriptionService) UpdateTask(hours ...int) error {
	hour, err := s.orFromPreferences(hours...)
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

	job, err := s.cronService.NewJob(
		gocron.CronJob(cronString, false),
		s.subscriptionTask(hour))

	if err != nil {
		return err
	}

	s.job = job
	return nil
}

func (s *subscriptionService) All() ([]models.Subscription, error) {
	return s.db.Subscriptions.All()
}

func (s *subscriptionService) Get(id uint) (*models.Subscription, error) {
	return s.db.Subscriptions.Get(id)
}

func (s *subscriptionService) Add(sub models.Subscription) (*models.Subscription, error) {
	existing, err := s.db.Subscriptions.GetByContentId(sub.ContentId)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("subscription already exists")
	}

	err = sub.Normalize(s.db.Preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize subscription: %w", err)
	}

	newSub, err := s.db.Subscriptions.New(sub)
	if err != nil {
		return nil, err
	}

	return newSub, nil
}

func (s *subscriptionService) Update(sub models.Subscription) error {
	pref, err := s.db.Preferences.Get()
	if err != nil {
		return err
	}

	if err = sub.Normalize(s.db.Preferences); err != nil {
		return fmt.Errorf("failed to normalize subscription: %w", err)
	}

	sub.Info.NextExecution = sub.NextExecution(pref.SubscriptionRefreshHour)

	return s.db.Subscriptions.Update(sub)
}

func (s *subscriptionService) Delete(id uint) error {
	return s.db.Subscriptions.Delete(id)
}

func (s *subscriptionService) subscriptionTask(hour int) gocron.Task {
	s.log.Debug().Int("hour", hour).Msg("creating subscription task")
	return gocron.NewTask(func() {
		subs, err := s.All()
		if err != nil {
			s.log.Error().Err(err).Msg("failed to get subscriptions")
			s.notifier.NotifyContentQ(s.transloco.GetTranslation("failed-to-run-subscriptions"),
				s.transloco.GetTranslation("failed-to-run-subscriptions-body", err), models.Red)
			return
		}

		counter := 0
		now := time.Now()
		for _, sub := range subs {
			nextExec := sub.NextExecution(hour)
			if !utils.IsSameDay(now, nextExec) {
				// Subscription only run once a day, if these don't match. It's for another day.
				continue
			}

			s.handleSub(sub, hour)
			counter++
		}

		s.log.Debug().Int("counter", counter).Msg("ran subscriptions")
	})
}

func (s *subscriptionService) handleSub(sub models.Subscription, hour int) {
	err := s.contentService.DownloadSubscription(&sub)
	sub.Info.LastCheck = time.Now()
	sub.Info.LastCheckSuccess = err == nil
	sub.Info.NextExecution = sub.NextExecution(hour)

	if err != nil {
		s.log.Error().Err(err).
			Uint("id", sub.ID).
			Str("contentId", sub.ContentId).
			Msg("failed to download content")
		s.notifier.NotifyContentQ(s.transloco.GetTranslation("failed-sub"),
			s.transloco.GetTranslation("failed-start-sub-download", sub.Info.Title, err))
		return
	}

	if err = s.db.Subscriptions.Update(sub); err != nil {
		s.log.Warn().Err(err).Uint("id", sub.ID).Msg("failed to update subscription")
	} else {
		s.log.Debug().Uint("id", sub.ID).Msg("updated subscription")
	}
}
