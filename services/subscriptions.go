package services

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
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
	// Update an existing subscription, updates DB, and restarts cron job
	// Subscription is normalized in the process
	Update(models.Subscription) error
	// Delete the subscription with ID
	Delete(uint) error
}

type subscriptionService struct {
	cronService    CronService
	contentService ContentService

	db  *db.Database
	log zerolog.Logger

	mapper  utils.SafeMap[uint, uuid.UUID]
	updator chan models.Subscription
}

func SubscriptionServiceProvider(db *db.Database, provider ContentService,
	log zerolog.Logger, cronService CronService) SubscriptionService {
	service := &subscriptionService{
		cronService:    cronService,
		contentService: provider,
		db:             db,
		log:            log.With().Str("handler", "subscription-service").Logger(),
		mapper:         utils.NewSafeMap[uint, uuid.UUID](),
		updator:        make(chan models.Subscription),
	}

	service.onStartUp()
	return service
}

func (s *subscriptionService) onStartUp() {
	go s.updateProcessor()

	subs, err := s.All()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get all subscriptions, cannot start jobs")
		return
	}

	failed := 0
	for _, sub := range subs {
		if err = s.schedule(sub); err != nil {
			failed++
			s.log.Error().Err(err).
				Uint("ID", sub.ID).
				Str("title", sub.Info.Title).
				Msg("Failed to schedule subscription")
		}
	}

	s.log.Info().Int("count", len(subs)-failed).Msg("scheduled subscriptions")
}

func (s *subscriptionService) updateProcessor() {
	for sub := range s.updator {
		err := s.db.Subscriptions.Update(sub)
		if err != nil {
			s.log.Warn().Err(err).Uint("id", sub.ID).Msg("failed to update subscription")
		} else {
			s.log.Debug().Uint("id", sub.ID).Msg("updated subscription")
		}
	}
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

	err = s.schedule(*newSub)
	if err != nil {
		return nil, err
	}

	return newSub, nil
}

func (s *subscriptionService) Update(sub models.Subscription) error {
	var existing *models.Subscription
	var err error

	err = sub.Normalize(s.db.Preferences)
	if err != nil {
		return fmt.Errorf("failed to normalize subscription: %w", err)
	}

	existing, err = s.db.Subscriptions.GetByContentId(sub.ContentId)
	if err != nil {
		return err
	}

	err = s.db.Subscriptions.Update(sub)
	if err != nil {
		return err
	}

	if existing != nil && !sub.ShouldRefresh(existing) {
		s.log.Debug().Uint("id", sub.ID).Msg("not refreshing subscription job")
		return nil
	}

	if ud, ok := s.mapper.Get(sub.ID); ok {
		if err = s.cronService.RemoveJob(ud); err != nil {
			s.log.Error().Err(err).Uint("id", sub.ID).Msg("failed to remove job")
			return err
		}
	} else {
		s.log.Trace().Err(err).Uint("id", sub.ID).Msg("updating subscription with no running job?")
	}

	err = s.schedule(sub)
	if err != nil {
		return err
	}

	return nil
}

func (s *subscriptionService) Delete(id uint) error {
	if ud, ok := s.mapper.Get(id); ok {
		if err := s.cronService.RemoveJob(ud); err != nil {
			s.log.Error().Err(err).Uint("id", id).Msg("failed to remove job")
			return err
		}

		s.mapper.Delete(id)
	}

	return s.db.Subscriptions.Delete(id)
}

// schedule the job for the passed subscription, optionally starting immediately
// Adds the job UUID to the mapper, does not save to DB.
func (s *subscriptionService) schedule(sub models.Subscription) error {
	nextExecutionTime, err := sub.NextExecution(s.db.Preferences)
	if err != nil {
		s.log.Error().Err(err).Uint("id", sub.ID).Msg("failed to get next execution")
		return err
	}

	job, err := s.cronService.NewJob(gocron.DurationJob(sub.RefreshFrequency.AsDuration()), s.toTask(sub),
		gocron.WithStartAt(gocron.WithStartDateTime(nextExecutionTime)))
	if err != nil {
		s.log.Error().Err(err).
			Uint("id", sub.ID).
			Time("nextExecution", nextExecutionTime).
			Msg("failed to create job")
		return fmt.Errorf("failed to start subscription job: %w", err)
	}

	s.mapper.Set(sub.ID, job.ID())
	s.log.Debug().
		Uint("id", sub.ID).
		Str("contentId", sub.ContentId).
		Str("title", sub.Info.Title).
		Time("nextExecution", nextExecutionTime).
		Dur("duration", sub.RefreshFrequency.AsDuration()).
		Msg("added subscription")
	return nil
}

func (s *subscriptionService) toTask(sub models.Subscription) gocron.Task {
	return gocron.NewTask(func() {
		err := s.contentService.DownloadSubscription(&sub)
		sub.Info.LastCheck = time.Now()
		sub.Info.LastCheckSuccess = err == nil

		if err != nil {
			s.log.Error().Err(err).
				Uint("id", sub.ID).
				Str("contentId", sub.ContentId).
				Msg("failed to download content")
			return
		}

		s.updator <- sub
	})
}
