package subscriptions

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"time"
)

var handler SubscriptionHandler

type SubscriptionHandler struct {
	db       *db.Database
	provider *providers.ContentProvider

	scheduler gocron.Scheduler
	idMapper  map[uint]uuid.UUID
	log       zerolog.Logger

	subUpdator chan models.Subscription
}

func New(db *db.Database, provider *providers.ContentProvider, log zerolog.Logger) (*SubscriptionHandler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	handler = SubscriptionHandler{
		scheduler:  s,
		db:         db,
		provider:   provider,
		idMapper:   make(map[uint]uuid.UUID),
		log:        log.With().Str("handler", "subscriptions").Logger(),
		subUpdator: make(chan models.Subscription, 100),
	}

	handler.initUpdateProcessor()

	handler.StartAll()
	handler.scheduler.Start()
	return &handler, nil
}

func (h *SubscriptionHandler) initUpdateProcessor() {
	go func() {
		for sub := range h.subUpdator {
			err := h.db.Subscriptions.Update(sub)
			if err != nil {
				h.log.Warn().Err(err).Uint("id", sub.ID).Msg("failed to update subscription")
			} else {
				h.log.Debug().Uint("id", sub.ID).Msg("updated subscription")
			}
		}
	}()
}

func Refresh(id uint) {
	handler.refresh(id)
}

func Delete(id uint) error {
	return handler.delete(id)
}

func (h *SubscriptionHandler) delete(id uint) error {
	mappedUuid, ok := h.idMapper[id]
	if !ok {
		return errors.New("subscription not found")
	}

	if err := h.scheduler.RemoveJob(mappedUuid); err != nil {
		h.log.Error().Err(err).Uint("id", id).Msg("failed to remove job")
		return err
	}

	delete(h.idMapper, id)
	h.log.Debug().Uint("id", id).Msg("removed subscription")
	return nil
}

func (h *SubscriptionHandler) refresh(id uint) {
	mappedUuid, ok := h.idMapper[id]
	if ok {
		if err := h.scheduler.RemoveJob(mappedUuid); err != nil {
			h.log.Error().Err(err).Uint("id", id).Msg("failed to remove job")
			return
		}
		h.log.Debug().Uint("id", id).Msg("removed subscription")
	}

	sub, err := h.db.Subscriptions.Get(id)
	if err != nil || sub == nil {
		h.log.Error().Err(err).Uint("id", id).Msg("failed to get subscription")
		return
	}

	h.new(*sub)
}

func (h *SubscriptionHandler) new(sub models.Subscription) {
	diff := time.Since(sub.Info.LastCheck)

	j, err := h.scheduler.NewJob(gocron.DurationJob(sub.RefreshFrequency.AsDuration()), h.toTask(sub),
		gocron.WithStartAt(func() gocron.StartAtOption {
			if diff > sub.RefreshFrequency.AsDuration() {
				h.log.Debug().
					Uint("id", sub.ID).
					Str("title", sub.Info.Title).
					Msg("subscription scheduled to execute immediately")
				return gocron.WithStartImmediately()
			}

			startTime := time.Now().Add(sub.RefreshFrequency.AsDuration() - diff)

			h.log.Debug().
				Uint("id", sub.ID).
				Str("title", sub.Info.Title).
				Time("at", startTime).
				Msg("subscription scheduled to execute")
			return gocron.WithStartDateTime(startTime)
		}()))

	if err != nil {
		h.log.Error().Err(err).Uint("id", sub.ID).Msg("failed to create job")
		return
	}

	h.idMapper[sub.ID] = j.ID()
	h.log.Debug().
		Uint("id", sub.ID).
		Str("contentId", sub.ContentId).
		Str("title", sub.Info.Title).
		Dur("duration", sub.RefreshFrequency.AsDuration()).
		Msg("added subscription")
}

func (h *SubscriptionHandler) StartAll() {
	subs, err := h.db.Subscriptions.All()
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get subscriptions")
		return
	}

	for _, sub := range subs {
		h.new(sub)
	}
	h.log.Info().Int("count", len(subs)).Msg("added subscriptions")
}

func (h *SubscriptionHandler) toTask(sub models.Subscription) gocron.Task {
	return gocron.NewTask(func() {
		err := h.provider.Download(payload.DownloadRequest{
			Id:        sub.ContentId,
			Provider:  sub.Provider,
			TempTitle: sub.Info.Title,
			BaseDir:   sub.Info.BaseDir,
		})
		sub.Info.LastCheck = time.Now()
		sub.Info.LastCheckSuccess = err == nil

		if err != nil {
			h.log.Error().Err(err).
				Uint("id", sub.ID).
				Str("contentId", sub.ContentId).
				Msg("failed to download content")
			return
		}

		h.subUpdator <- sub
	})
}
