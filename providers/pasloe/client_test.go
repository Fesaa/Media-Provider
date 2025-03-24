package pasloe

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/utils/mock"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"net/http"
	"testing"
)

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func testClient(t *testing.T) api.Client {
	t.Helper()

	cont := dig.New()

	must(t, cont.Provide(utils.Identity(afero.Afero{Fs: afero.NewMemMapFs()})))
	must(t, cont.Provide(utils.Identity(&config.Config{})))
	must(t, cont.Provide(utils.Identity(http.DefaultClient)))
	must(t, cont.Provide(utils.Identity(cont)))
	must(t, cont.Provide(utils.Identity(zerolog.Nop())))
	must(t, cont.Provide(func() services.SignalRService { return &mock.SignalR{} }))
	must(t, cont.Provide(func() services.NotificationService { return &mock.Notifications{} }))
	must(t, cont.Provide(func() models.Preferences { return &mock.Preferences{} }))
	must(t, cont.Provide(func() services.TranslocoService { return &mock.Transloco{} }))
	must(t, cont.Provide(func() services.CacheService { return &mock.Cache{} }))
	must(t, cont.Provide(mangadex.NewRepository))
	must(t, cont.Provide(dynasty.NewRepository))
	must(t, cont.Provide(webtoon.NewRepository))
	must(t, cont.Provide(services.DirectoryServiceProvider))
	must(t, cont.Provide(services.MarkdownServiceProvider))
	must(t, cont.Provide(services.ImageServiceProvider))
	must(t, cont.Provide(New))
	c := utils.MustInvokeCont[api.Client](cont).(*client)

	c.registry = &mockRegistry{}

	return c
}

type stateInfo struct {
	state    payload.ContentState
	provider models.Provider
	callback func(api.Downloadable)
}

func (si stateInfo) ToContent(id string) api.Downloadable {
	return &MockContent{
		mockProvider: si.provider,
		mockState:    si.state,
		mockId:       id,
	}
}

func setupQueue(t *testing.T, pasloe *client, infos ...stateInfo) (content []api.Downloadable) {
	t.Helper()

	for i, info := range infos {
		id := fmt.Sprintf("%d", i)
		c := info.ToContent(id)
		pasloe.content.Set(id, c)
		content = append(content, c)
	}
	return
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

			got := c.CanStart(tc.provider)
			if got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}

func TestClient_QueueTests(t *testing.T) {
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
			},
			want: payload.ContentStateDownloading,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t)

			if len(tt.current) > 0 {
				setupQueue(t, c.(*client), tt.current...)
			}

			if err := c.Download(tt.enqueue); err != nil {
				t.Fatal(err)
			}

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
