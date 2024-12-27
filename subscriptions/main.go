package subscriptions

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

var handler subscriptionHandler

type subscriptionHandler struct {
	scheduler gocron.Scheduler
	db        *db.Database
	idMapper  map[int64]uuid.UUID
	log       *log.Logger

	subUpdator chan models.Subscription
}

func Init(db *db.Database) {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal("Failed to initialize scheduler", err)
	}

	handler = subscriptionHandler{
		scheduler:  s,
		db:         db,
		idMapper:   make(map[int64]uuid.UUID),
		log:        log.With("handler", "subscriptions"),
		subUpdator: make(chan models.Subscription, 100),
	}

	handler.initUpdateProcessor()

	handler.StartAll()
	handler.scheduler.Start()
}

func (h *subscriptionHandler) initUpdateProcessor() {
	go func() {
		for sub := range h.subUpdator {
			err := h.db.Subscriptions.Update(sub)
			if err != nil {
				h.log.Warn("Error updating subscription check time",
					"id", sub.Id, "err", err)
			} else {
				h.log.Info("Subscription updated successfully",
					"id", sub.Id, "lastCheck", sub.Info.LastCheck)
			}
		}
	}()
}

func Refresh(id int64) {
	handler.refresh(id)
}

func Delete(id int64) error {
	return handler.delete(id)
}

func (h *subscriptionHandler) delete(id int64) error {
	mappedUuid, ok := h.idMapper[id]
	if !ok {
		return errors.New("subscription not found")
	}

	if err := h.scheduler.RemoveJob(mappedUuid); err != nil {
		h.log.Error("Failed to remove job", "id", mappedUuid, "err", err)
		return err
	}

	delete(h.idMapper, id)
	h.log.Info("Removed job", "id", mappedUuid, "subscriptionId", id)
	return nil
}

func (h *subscriptionHandler) refresh(id int64) {
	mappedUuid, ok := h.idMapper[id]
	if ok {
		if err := h.scheduler.RemoveJob(mappedUuid); err != nil {
			h.log.Error("Failed to remove job", "id", mappedUuid, "err", err)
			return
		}
		h.log.Debug("Removed job", "id", mappedUuid, "subscriptionId", id)
	}

	sub, err := h.db.Subscriptions.Get(id)
	if err != nil || sub == nil {
		h.log.Error("Failed to get subscription", "id", id, "err", err)
		return
	}

	h.new(*sub)
}

func (h *subscriptionHandler) new(sub models.Subscription) {
	diff := time.Since(sub.Info.LastCheck)

	j, err := h.scheduler.NewJob(gocron.DurationJob(sub.RefreshFrequency.AsDuration()), h.toTask(sub),
		gocron.WithStartAt(func() gocron.StartAtOption {
			if diff > sub.RefreshFrequency.AsDuration() {
				h.log.Debug("subscription scheduled to execute immediately", "id", sub.Id)
				return gocron.WithStartImmediately()
			}

			startTime := time.Now().Add(sub.RefreshFrequency.AsDuration() - diff)

			h.log.Debug("subscription scheduled to execute", "id", sub.Id, "title", sub.Info.Title, slog.Time("at", startTime))
			return gocron.WithStartDateTime(startTime)
		}()))

	if err != nil {
		h.log.Error("Error creating subscription job", "id", sub.Id)
		return
	}

	h.idMapper[sub.Id] = j.ID()
	h.log.Debug("Subscription scheduled",
		"subId", sub.Id, "contentId", sub.ContentId, "uuid", j.ID(), "title", sub.Info.Title,
		slog.Duration("duration", sub.RefreshFrequency.AsDuration()))
}

func (h *subscriptionHandler) StartAll() {
	subs, err := h.db.Subscriptions.All()
	if err != nil {
		h.log.Error("Error getting all subscriptions, cannot start cron jobs", "error", err)
		return
	}

	for _, sub := range subs {
		h.new(sub)
	}
	h.log.Info("Subscriptions loaded from database, and scheduled", slog.Int("amount", len(subs)))
}

func (h *subscriptionHandler) toTask(sub models.Subscription) gocron.Task {
	return gocron.NewTask(func() {
		err := providers.Download(payload.DownloadRequest{
			Id:        sub.ContentId,
			Provider:  sub.Provider,
			TempTitle: sub.Info.Title,
			BaseDir:   sub.Info.BaseDir,
		})
		sub.Info.LastCheck = time.Now()
		sub.Info.LastCheckSuccess = err == nil

		if err != nil {
			h.log.Error("Error downloading subscription, check config",
				"id", sub.Id,
				"contentId", sub.ContentId,
				"error", err)
			return
		}

		h.subUpdator <- sub
	})
}
