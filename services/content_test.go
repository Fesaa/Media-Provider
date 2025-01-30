package services

import (
	"bytes"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
	"strings"
	"testing"
	"time"
)

func tempContentService(t *testing.T) (*db.Database, ContentService) {
	t.Helper()

	log := zerolog.Nop()

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, err := db.DatabaseProvider(log)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.RootDir = tempDir

	cs := ContentServiceProvider(log)
	cs.RegisterProvider(models.LIME, providerAdapterMock{})
	cs.RegisterProvider(models.YTS, providerAdapterMock{})
	cs.RegisterProvider(models.NYAA, providerAdapterMock{})
	cs.RegisterProvider(models.MANGADEX, providerAdapterMock{})
	cs.RegisterProvider(models.SUBSPLEASE, providerAdapterMock{})
	cs.RegisterProvider(models.DYNASTY, providerAdapterMock{})
	cs.RegisterProvider(models.WEBTOON, providerAdapterMock{})

	return database, cs
}

func TestContentService_Search(t *testing.T) {
	t.Parallel()
	_, cs := tempContentService(t)

	req := payload.SearchRequest{
		Provider:  []models.Provider{models.MANGADEX},
		Query:     "Spice And Wolf",
		Modifiers: nil,
	}

	_, err := cs.Search(req)
	if err != nil {
		t.Fatal(err)
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

type mockClient struct {
}

func (m mockClient) Download(request payload.DownloadRequest) error {
	return nil
}

func (m mockClient) RemoveDownload(request payload.StopRequest) error {
	return nil
}

func (m mockClient) Content(s string) Content {
	return nil
}

type providerAdapterMock struct {
}

func (p providerAdapterMock) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{}
}

func (p providerAdapterMock) Search(request payload.SearchRequest) ([]payload.Info, error) {
	return []payload.Info{}, nil
}

func (p providerAdapterMock) Client() Client {
	return mockClient{}
}

type slowBuilder struct{}

func (slowBuilder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{}
}

func (slowBuilder) Search(request payload.SearchRequest) ([]payload.Info, error) {
	time.Sleep(3 * time.Second)
	return []payload.Info{}, nil
}

func (p slowBuilder) Client() Client {
	return slowMockClient{}
}

type slowMockClient struct{}

func (s slowMockClient) Download(request payload.DownloadRequest) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (s slowMockClient) RemoveDownload(request payload.StopRequest) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (s slowMockClient) Content(s2 string) Content {
	return nil
}
