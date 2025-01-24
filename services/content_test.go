package services

import (
	"bytes"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"net/http"
	"strings"
	"testing"
	"time"
)

func tempContentService(t *testing.T) (*db.Database, ContentService) {
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	log := zerolog.New(zerolog.NewConsoleWriter())

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, err := db.DatabaseProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.RootDir = tempDir

	cont := dig.New()
	must(cont.Provide(utils.Identity(log)))
	must(cont.Provide(utils.Identity(database)))
	must(cont.Provide(utils.Identity(http.DefaultClient)))
	must(cont.Provide(utils.Identity(cfg)))
	must(cont.Provide(mangadex.NewRepository))
	must(cont.Provide(dynasty.NewRepository))
	must(cont.Provide(yoitsu.New))
	must(cont.Provide(pasloe.New))
	must(cont.Provide(utils.Identity(cont)))

	return database, ContentServiceProvider(cont, log)
}

func TestContentService_Search(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)

	req := payload.SearchRequest{
		Provider:  []models.Provider{models.MANGADEX},
		Query:     "Spice And Wolf",
		Modifiers: nil,
	}

	data, err := cs.Search(req)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("Empty result")
	}
}

func TestContentService_SearchInvalidProvider(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)
	req := payload.SearchRequest{
		Provider:  []models.Provider{models.Provider(9999)},
		Query:     "Spice And Wolf",
		Modifiers: nil,
	}

	_, err := cs.Search(req)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestContentService_DownloadAndStop(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)

	req := payload.DownloadRequest{
		Provider:  models.MANGADEX,
		Id:        "de900fd3-c94c-4148-bbcb-ca56eaeb57a4",
		BaseDir:   "Manga",
		TempTitle: "Spice And Wolf",
	}

	if err := cs.Download(req); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	stop := payload.StopRequest{
		Provider:    models.MANGADEX,
		Id:          "de900fd3-c94c-4148-bbcb-ca56eaeb57a4",
		DeleteFiles: true,
	}

	if err := cs.Stop(stop); err != nil {
		t.Fatal(err)
	}
}

func TestContentService_DownloadInvalid(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)
	req := payload.DownloadRequest{
		Provider: models.Provider(999),
	}

	if err := cs.Download(req); err == nil {
		t.Fatal("Should have errored")
	}
}

func TestContentService_StopInvalid(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)
	req := payload.StopRequest{
		Provider: models.Provider(999),
	}

	if err := cs.Stop(req); err == nil {
		t.Fatal("Should have errored")
	}
}

func TestContentService_DownloadSub(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)
	sub := models.Subscription{
		Provider:  models.MANGADEX,
		ContentId: "de900fd3-c94c-4148-bbcb-ca56eaeb57a4",
		Info: models.SubscriptionInfo{
			BaseDir: "Manga",
			Title:   "Spice And Wolf",
		},
	}

	if err := cs.DownloadSubscription(&sub); err != nil {
		t.Fatal(err)
	}
}

func TestContentService_SearchInvalidSearchOptiosn(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)
	req := payload.SearchRequest{
		Provider: []models.Provider{models.MANGADEX},
		Query:    "Spice And Wolf",
		Modifiers: map[string][]string{
			"SkipNotFoundTags": {"false"},
			"includeTags":      {"Random Not Matching Tag"},
		},
	}

	_, err := cs.Search(req)
	if err == nil {
		t.Fatal("Should have errored")
	}
}

func TestContentService_SearchVerySlow(t *testing.T) {
	t.Parallel()
	_, cst := tempContentService(t)
	cs := cst.(*contentService)

	var logBuffer bytes.Buffer
	cs.log = zerolog.New(&logBuffer)

	cs.providers.Set(models.Provider(999), &slowBuilder{})

	req := payload.SearchRequest{
		Provider:  []models.Provider{models.Provider(999)},
		Query:     "Spice And Wolf",
		Modifiers: nil,
	}

	_, err := cs.Search(req)
	if err != nil {
		t.Fatal(err)
	}

	log := logBuffer.String()

	if !strings.Contains(log, "searching took more than one second") {
		t.Fatalf("Log should contain \"searching took more than one second\", got \n\n%s", log)
	}

}

type slowBuilder struct{}

func (slowBuilder) Search(request payload.SearchRequest) ([]payload.Info, error) {
	time.Sleep(3 * time.Second)
	return []payload.Info{}, nil
}

func (slowBuilder) Download(request payload.DownloadRequest) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (slowBuilder) Stop(request payload.StopRequest) error {
	time.Sleep(3 * time.Second)
	return nil
}
