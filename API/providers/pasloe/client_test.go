package pasloe

import (
	"fmt"
	"testing"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/utils/mock"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func testClient(t *testing.T, options ...utils.Option[*client]) core.Client {
	t.Helper()

	cont := dig.New()

	must(t, cont.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))
	must(t, cont.Provide(utils.Identity(&config.Config{})))
	must(t, cont.Provide(utils.Identity(menou.DefaultClient)))
	must(t, cont.Provide(utils.Identity(cont)))
	must(t, cont.Provide(utils.Identity(zerolog.Nop())))
	must(t, cont.Provide(func() services.SignalRService { return &mock.SignalR{} }))
	must(t, cont.Provide(func() services.NotificationService { return &mock.Notifications{} }))
	must(t, cont.Provide(func() services.TranslocoService { return &mock.Transloco{} }))
	must(t, cont.Provide(func() services.CacheService { return &mock.Cache{} }))
	must(t, cont.Provide(func() services.SettingsService { return &mock.Settings{} }))
	must(t, cont.Provide(func() *db.UnitOfWork { return nil }))
	must(t, cont.Provide(mangadex.NewRepository))
	must(t, cont.Provide(dynasty.NewRepository))
	must(t, cont.Provide(webtoon.NewRepository))
	must(t, cont.Provide(services.DirectoryServiceProvider))
	must(t, cont.Provide(services.MarkdownServiceProvider))
	must(t, cont.Provide(services.ImageServiceProvider))
	must(t, cont.Provide(services.ArchiveServiceProvider))
	must(t, cont.Provide(New))
	c := utils.MustInvoke[core.Client](cont).(*client)

	c.registry = &mockRegistry{
		cont: cont,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

type stateInfo struct {
	state    payload.ContentState
	provider models.Provider
	callback func(core.Downloadable)
}

func (si stateInfo) ToContent(id string) core.Downloadable {
	return &MockContent{
		mockProvider: si.provider,
		mockId:       id,
	}
}

func setupQueue(t *testing.T, pasloe *client, infos ...stateInfo) {
	t.Helper()

	for i, info := range infos {
		id := fmt.Sprintf("%d", i)

		c, _ := pasloe.registry.Create(pasloe, payload.DownloadRequest{
			Provider: info.provider,
			Id:       id,
		})
		c.SetState(info.state)

		pasloe.content.Set(id, c)
	}
}

func TestClient_CanStart(t *testing.T) {
	type testCase struct {
		name     string
		current  []stateInfo
		provider models.Provider
		want     bool
	}

	testCases := []testCase{
		{
			name:     "Empty queue",
			current:  nil,
			provider: models.MANGADEX,
			want:     true,
		},
		{
			name: "Some already downloading",
			current: []stateInfo{
				{
					state:    payload.ContentStateDownloading,
					provider: models.MANGADEX,
				},
			},
			provider: models.MANGADEX,
			want:     false,
		},
		{
			name: "Some queued",
			current: []stateInfo{
				{
					state:    payload.ContentStateQueued,
					provider: models.MANGADEX,
				},
			},
			provider: models.MANGADEX,
			want:     true,
		},
		{
			name: "Some loading info",
			current: []stateInfo{
				{
					state:    payload.ContentStateLoading,
					provider: models.MANGADEX,
				},
			},
			provider: models.MANGADEX,
			want:     false,
		},
		{
			name: "Some cleanup",
			current: []stateInfo{
				{
					state:    payload.ContentStateCleanup,
					provider: models.MANGADEX,
				},
			},
			provider: models.MANGADEX,
			want:     true,
		},
		{
			name: "Some waiting",
			current: []stateInfo{
				{
					state:    payload.ContentStateWaiting,
					provider: models.MANGADEX,
				},
			},
			provider: models.MANGADEX,
			want:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(t)

			if len(tc.current) > 0 {
				setupQueue(t, c.(*client), tc.current...)
			}

		})
	}
}

func TestClient_QueueTests(t *testing.T) {
	t.Skipf("Test skipped as underlying content has become too smart to mock like this")

	type queueTest struct {
		name    string
		current []stateInfo
		enqueue payload.DownloadRequest
		want    payload.ContentState
	}

	tests := []queueTest{
		{
			name:    "Empty Queue Add",
			current: nil,
			enqueue: payload.DownloadRequest{
				Provider:  models.MANGADEX,
				Id:        "MyID",
				BaseDir:   "Manga",
				TempTitle: "Spice and Wolf",
				DownloadMetadata: models.DownloadRequestMetadata{
					StartImmediately: true,
				},
			},
			want: payload.ContentStateDownloading,
		},
		{
			name: "Busy",
			current: []stateInfo{
				{
					state:    payload.ContentStateLoading,
					provider: models.MANGADEX,
				},
			},
			enqueue: payload.DownloadRequest{
				Provider:  models.MANGADEX,
				Id:        "MyID",
				BaseDir:   "Manga",
				TempTitle: "Spice and Wolf",
				DownloadMetadata: models.DownloadRequestMetadata{
					StartImmediately: true,
				},
			},
			want: payload.ContentStateQueued,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t, func(c *client) {
				c.registry.(*mockRegistry).finishContent = false
			})

			if len(tt.current) > 0 {
				setupQueue(t, c.(*client), tt.current...)
			}

			if err := c.Download(tt.enqueue); err != nil {
				t.Fatal(err)
			}

			time.Sleep(time.Millisecond * 10)

			content := c.Content(tt.enqueue.Id)
			if content == nil {
				t.Fatal("content is nil")
			}

			got := content.State()
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_QueueProgressing(t *testing.T) {
	t.Skipf("Test skipped as underlying content has become too smart to mock like this")

	c := testClient(t, func(c *client) {
		c.registry.(*mockRegistry).finishContent = false
	}).(*client)

	spiceAndWolf, _ := c.registry.Create(c, payload.DownloadRequest{
		Provider: models.MANGADEX,
		Id:       "Spice and Wolf",
	})

	theExecutionerAndHerWayOfLife, _ := c.registry.Create(c, payload.DownloadRequest{
		Provider: models.MANGADEX,
		Id:       "The Executioner and Her Way of Life",
	})

	// Setup state and some items to download
	spiceAndWolf.SetState(payload.ContentStateDownloading)
	spiceAndWolf.(*MockContent).ToDownload = []ID{"a", "b"}
	theExecutionerAndHerWayOfLife.SetState(payload.ContentStateReady)
	theExecutionerAndHerWayOfLife.(*MockContent).ToDownload = []ID{"a", "b"}
	c.content.Set(spiceAndWolf.Id(), spiceAndWolf)
	c.content.Set(theExecutionerAndHerWayOfLife.Id(), theExecutionerAndHerWayOfLife)

	// Add a new piece of content naturally
	if err := c.Download(payload.DownloadRequest{
		Provider: models.MANGADEX,
		Id:       "Otherside Picnic",
		DownloadMetadata: models.DownloadRequestMetadata{
			StartImmediately: false,
		},
	}); err != nil {
		t.Fatal(err)
	}

	othersidePicnic, _ := c.content.Get("Otherside Picnic")

	othersidePicnic.(*MockContent).loadInfoFunc = func() {
		// Don't close channel
	}

	// Queued as Spice and Wolf is downloaded
	got := othersidePicnic.State()
	if got != payload.ContentStateQueued {
		t.Fatalf("got %v, want %v", got, payload.ContentStateQueued)
	}

	if err := c.RemoveDownload(payload.StopRequest{
		Provider: models.MANGADEX,
		Id:       spiceAndWolf.Id(),
	}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 10)

	got = othersidePicnic.State()
	if got != payload.ContentStateLoading {
		t.Fatalf("got %v, want %v", got, payload.ContentStateLoading)
	}

	// Otherside Picnic is loading, cannot start yet
	got = theExecutionerAndHerWayOfLife.State()
	if got != payload.ContentStateReady {
		t.Fatalf("got %v, want %v", got, payload.ContentStateReady)
	}

	// Stop Otherside Picnic loading info
	close(othersidePicnic.(*MockContent).loadInfoChan)

	time.Sleep(time.Millisecond * 10)

	got = theExecutionerAndHerWayOfLife.State()
	if got != payload.ContentStateDownloading {
		t.Fatalf("got %v, want %v", got, payload.ContentStateDownloading)
	}

	got = othersidePicnic.State()
	if got != payload.ContentStateWaiting {
		t.Fatalf("got %v, want %v", got, payload.ContentStateWaiting)
	}

	if err := c.RemoveDownload(payload.StopRequest{
		Provider: models.MANGADEX,
		Id:       theExecutionerAndHerWayOfLife.Id(),
	}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 10)

	if _, err := othersidePicnic.Message(payload.Message{
		ContentId:   othersidePicnic.Id(),
		MessageType: payload.StartDownload,
	}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 10)

	got = othersidePicnic.State()
	if got != payload.ContentStateDownloading {
		t.Fatalf("got %v, want %v", got, payload.ContentStateDownloading)
	}

}
