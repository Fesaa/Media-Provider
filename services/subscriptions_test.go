package services

import (
	"bytes"
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"testing"
	"time"
)

func tempDatabase(t *testing.T) *db.Database {
	t.Helper()
	config.Dir = t.TempDir()
	//nolint:usetesting
	_ = os.Setenv("DATABASE_DSN", "file::memory:")
	database, err := db.DatabaseProvider(zerolog.Nop())
	if err != nil {
		if strings.Contains(err.Error(), "attempt to write a readonly database") {
			t.Skipf("ReadOnly DB error, I don't have a good way to fix this atm.")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d, err := database.DB().DB()
		if err != nil {
			t.Fatal(err)
		}
		d.Close()
	})
	return database
}

func tempSubscriptionService(t *testing.T, tempdb *db.Database, logs ...zerolog.Logger) SubscriptionService {
	t.Helper()
	log := utils.OrDefault(logs, zerolog.Nop())

	tempDir := t.TempDir()
	config.Dir = tempDir

	cs := tempContentService(t)
	cron, err := CronServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	transloco, err := TranslocoServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	signalR := SignalRServiceProvider(SignalRParams{
		Log: log,
	})

	return SubscriptionServiceProvider(tempdb, cs, log, cron, NotificationServiceProvider(log, tempdb, signalR), transloco)
}

func defaultSub() models.Subscription {
	return models.Subscription{
		Provider:         models.MANGADEX,
		ContentId:        "de900fd3-c94c-4148-bbcb-ca56eaeb57a4",
		RefreshFrequency: models.Day,
		Info: models.SubscriptionInfo{
			Title:            "Spice and Wolf",
			BaseDir:          "Manga",
			LastCheck:        time.Now(),
			LastCheckSuccess: true,
		},
	}
}

func TestSubscriptionService_All(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))

	sub := defaultSub()

	_, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}
	subs, err := ss.All()
	if err != nil {
		t.Fatal(err)
	}

	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription got %d", len(subs))
	}

}

func TestSubscriptionService_Add(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))
	sub := defaultSub()
	s, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ss.Get(s.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionService_AddDupe(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))

	sub := defaultSub()
	_, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ss.Add(sub)
	if err == nil {
		t.Fatal("Should not be able to add duplicate subscription")
	}

}

type brokenPreferences struct {
	failAfter int
	counter   int
}

func (b *brokenPreferences) Get() (*models.Preference, error) {
	if b.counter < b.failAfter {
		b.counter++
		return &models.Preference{
			SubscriptionRefreshHour: 0,
		}, nil
	}
	return nil, errors.New("broken preferences")
}

func (b *brokenPreferences) GetComplete() (*models.Preference, error) {
	return b.Get()
}

func (b *brokenPreferences) Update(pref models.Preference) error {
	return errors.New("broken preferences")
}

func TestSubscriptionService_AddBadPreference(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))

	ssImpl := ss.(*subscriptionService)
	ssImpl.db.Preferences = &brokenPreferences{
		failAfter: 0,
		counter:   0,
	}

	sub := defaultSub()

	_, err := ss.Add(sub)
	if !errors.Is(err, models.ErrFailedToLoadPreferences) {
		t.Fatalf("Wanted ErrFailedToLoadPreferences, got %v", err)
	}

	ssImpl.db.Preferences = &brokenPreferences{
		failAfter: 1,
		counter:   0,
	}

	sub = defaultSub()
	sub.ContentId = "7546ff2d-2310-47a4-b1f3-1a2561f20ce7"
	_, err = ss.Add(sub)
	if !errors.Is(err, models.ErrFailedToLoadPreferences) {
		t.Fatalf("Wanted ErrFailedToLoadPreferences, got %v", err)
	}
}

func TestSubscriptionService_UpdateNoRefresh(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))

	ssImpl := ss.(*subscriptionService)
	var buffer bytes.Buffer
	ssImpl.log = zerolog.New(&buffer)

	sub := defaultSub()
	_, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	sub.Info.Title = "Something New"
	if err = ss.Update(sub); err != nil {
		t.Fatal(err)
	}

	log := buffer.String()
	if !strings.Contains(log, "not refreshing subscription job") {
		t.Fatalf("Wanted not refreshing subscription job, got %s", log)
	}
}

func TestSubscriptionService_UpdateBadPreference(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))
	ssImpl := ss.(*subscriptionService)
	sub := defaultSub()
	_, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	ssImpl.db.Preferences = &brokenPreferences{}
	err = ss.Update(sub)
	if !errors.Is(err, models.ErrFailedToLoadPreferences) {
		t.Fatalf("Wanted ErrFailedToLoadPreferences, got %v", err)
	}
}

func TestSubscriptionService_UpdateRefresh(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))
	sub := defaultSub()
	_, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	sub.Info.BaseDir = "LightNovels"
	err = ss.Update(sub)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionService_toTask(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))
	sub := defaultSub()

	ssImpl := ss.(*subscriptionService)

	task := ssImpl.toTask(sub.ID)
	_, err := ssImpl.cronService.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()), task)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	// Cleanup
	err = ssImpl.contentService.Stop(payload.StopRequest{
		Provider:    sub.Provider,
		Id:          sub.ContentId,
		DeleteFiles: true,
	})
	if err != nil {
		t.Fatal(err)
	}

}

func TestSubscriptionService_toTaskFailedDownload(t *testing.T) {
	t.Parallel()
	sub := models.Subscription{
		Provider:         models.Provider(999),
		ContentId:        "RTFYGUHIJ",
		RefreshFrequency: models.Day,
		Info: models.SubscriptionInfo{
			Title:            "RTFYTGUHUJ",
			BaseDir:          "Manga",
			LastCheck:        time.Now(),
			LastCheckSuccess: true,
		},
	}
	var buffer bytes.Buffer
	log := zerolog.New(&buffer)

	ss := tempSubscriptionService(t, tempDatabase(t), log)
	ssImpl := ss.(*subscriptionService)

	task := ssImpl.toTask(sub.ID)
	_, err := ssImpl.cronService.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()), task)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure buffer has enough time to write log line
	time.Sleep(1 * time.Second)

	Log := buffer.String()
	if !strings.Contains(Log, "failed to get subscription") {
		t.Fatalf("Wanted failed to get subscription, got %s", Log)
	}

}

func TestSubscriptionServiceProvider_StartUp(t *testing.T) {
	t.Parallel()
	sub := defaultSub()
	database := tempDatabase(t)
	_, err := database.Subscriptions.New(sub)
	if err != nil {
		t.Fatal(err)
	}

	var buffer bytes.Buffer
	log := zerolog.New(&buffer)

	_ = tempSubscriptionService(t, database, log)

	Log := buffer.String()
	if !strings.Contains(Log, "scheduled subscriptions") ||
		!strings.Contains(Log, `"count":1`) {
		t.Fatalf("Wanted scheduled subscriptions, got %s", Log)
	}
}

func TestSubscriptionServiceProvider_FailAtStartUp(t *testing.T) {
	t.Parallel()

	sub := defaultSub()
	database := tempDatabase(t)
	_, err := database.Subscriptions.New(sub)
	if err != nil {
		t.Fatal(err)
	}

	database.Preferences = &brokenPreferences{}

	var buffer bytes.Buffer
	log := zerolog.New(&buffer)
	_ = tempSubscriptionService(t, database, log)

	Log := buffer.String()
	if !strings.Contains(Log, "Failed to schedule subscription") ||
		!strings.Contains(Log, `"error":"failed to load preferences"`) {
		t.Fatalf("Wanted scheduled subscriptions, got %s", Log)
	}
}

func TestSubscriptionServiceProvider_Delete(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t, tempDatabase(t))
	sub := defaultSub()

	newSub, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	if err = ss.Delete(newSub.ID); err != nil {
		t.Fatal(err)
	}
}
