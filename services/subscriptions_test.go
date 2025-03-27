package services

import (
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
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

func tempSubscriptionService(t *testing.T, tempdb *db.Database) SubscriptionService {
	t.Helper()
	log := zerolog.Nop()

	tempDir := t.TempDir()
	config.Dir = tempDir

	cs := tempContentService(t)
	cron, err := CronServiceProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	transloco, err := TranslocoServiceProvider(log, afero.Afero{Fs: afero.NewMemMapFs()})
	if err != nil {
		t.Fatal(err)
	}

	signalR := SignalRServiceProvider(SignalRParams{
		Log: log,
	})

	ss, err := SubscriptionServiceProvider(tempdb, cs, log, cron, NotificationServiceProvider(log, tempdb, signalR), transloco)
	if err != nil {
		t.Fatal(err)
	}
	return ss
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
