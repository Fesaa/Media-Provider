package services

import (
	"bytes"
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"strings"
	"testing"
	"time"
)

func tempSubscriptionService(t *testing.T) SubscriptionService {
	t.Helper()
	log := zerolog.Nop()

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, cs := tempContentService(t)
	cron, err := CronServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	return SubscriptionServiceProvider(database, cs, log, cron)
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
	ss := tempSubscriptionService(t)

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
	ss := tempSubscriptionService(t)
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
	ss := tempSubscriptionService(t)

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

func (b *brokenPreferences) Update(pref models.Preference) error {
	return errors.New("broken preferences")
}

func TestSubscriptionService_AddBadPreference(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t)

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
	ss := tempSubscriptionService(t)

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
	ss := tempSubscriptionService(t)
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
	ss := tempSubscriptionService(t)
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
	ss := tempSubscriptionService(t)
	sub := defaultSub()

	ssImpl := ss.(*subscriptionService)

	task := ssImpl.toTask(sub)
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

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, cs := tempContentService(t)
	cron, err := CronServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	ss := SubscriptionServiceProvider(database, cs, log, cron)
	ssImpl := ss.(*subscriptionService)

	task := ssImpl.toTask(sub)
	_, err = ssImpl.cronService.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()), task)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure buffer has enough time to write log line
	time.Sleep(1 * time.Second)

	Log := buffer.String()
	if !strings.Contains(Log, "failed to download content") {
		t.Fatalf("Wanted failed to download content, got %s", Log)
	}

}

func TestSubscriptionServiceProvider_StartUp(t *testing.T) {
	t.Parallel()
	log := zerolog.New(zerolog.NewConsoleWriter())

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, cs := tempContentService(t)
	cron, err := CronServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	sub := defaultSub()
	_, err = database.Subscriptions.New(sub)
	if err != nil {
		t.Fatal(err)
	}

	var buffer bytes.Buffer
	log = zerolog.New(&buffer)

	_ = SubscriptionServiceProvider(database, cs, log, cron)

	Log := buffer.String()
	if !strings.Contains(Log, "scheduled subscriptions") ||
		!strings.Contains(Log, `"count":1`) {
		t.Fatalf("Wanted scheduled subscriptions, got %s", Log)
	}
}

func TestSubscriptionServiceProvider_FailAtStartUp(t *testing.T) {
	t.Parallel()
	log := zerolog.New(zerolog.NewConsoleWriter())

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, cs := tempContentService(t)
	cron, err := CronServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	sub := defaultSub()
	_, err = database.Subscriptions.New(sub)
	if err != nil {
		t.Fatal(err)
	}

	database.Preferences = &brokenPreferences{}

	var buffer bytes.Buffer
	log = zerolog.New(&buffer)
	_ = SubscriptionServiceProvider(database, cs, log, cron)

	Log := buffer.String()
	if !strings.Contains(Log, "Failed to schedule subscription") ||
		!strings.Contains(Log, `"error":"failed to load preferences"`) {
		t.Fatalf("Wanted scheduled subscriptions, got %s", Log)
	}
}

func TestSubscriptionServiceProvider_Delete(t *testing.T) {
	t.Parallel()
	ss := tempSubscriptionService(t)
	sub := defaultSub()

	newSub, err := ss.Add(sub)
	if err != nil {
		t.Fatal(err)
	}

	if err = ss.Delete(newSub.ID); err != nil {
		t.Fatal(err)
	}
}
