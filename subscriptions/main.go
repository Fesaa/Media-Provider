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
)

var handler subscriptionHandler

type subscriptionHandler struct {
	scheduler gocron.Scheduler
	db        *db.Database
	idMapper  map[int64]uuid.UUID
	log       *log.Logger
}

func Init(db *db.Database) {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal("Failed to initialize scheduler", err)
	}

	handler = subscriptionHandler{
		scheduler: s,
		db:        db,
		idMapper:  make(map[int64]uuid.UUID),
		log:       log.With("handler", "subscriptions"),
	}

	handler.StartAll()
	handler.scheduler.Start()
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
	}

	sub, err := h.db.Subscriptions.Get(id)
	if err != nil || sub == nil {
		h.log.Error("Failed to get subscription", "id", id, "err", err)
		return
	}

	h.new(*sub)
}

func (h *subscriptionHandler) new(sub models.Subscription) {
	j, err := h.scheduler.NewJob(gocron.DurationJob(sub.RefreshFrequency.AsDuration()), h.toTask(sub))

	if err != nil {
		h.log.Error("Error creating subscription job", "id", sub.Id)
		return
	}

	h.idMapper[sub.Id] = j.ID()
	h.log.Info("Subscription scheduled",
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
}

func (h *subscriptionHandler) toTask(sub models.Subscription) gocron.Task {
	return gocron.NewTask(func() {
		err := providers.Download(payload.DownloadRequest{
			Id:        sub.ContentId,
			Provider:  sub.Provider,
			TempTitle: sub.Info.Title,
			BaseDir:   sub.Info.BaseDir,
		})
		if err != nil {
			h.log.Error("Error downloading subscription, check config",
				"id", sub.Id,
				"contentId", sub.ContentId,
				"error", err)
		}
	})
}
