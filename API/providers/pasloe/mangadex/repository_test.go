package mangadex

import (
	"context"
	"io"
	"testing"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils/mock"
	"github.com/rs/zerolog"
)

// Note: Sleeping quite a bit during test, to ensure we do not send
// too many requests to MangeDex
// Do NOT run these in parallel
//
// Repository.GetCoverImages's offset traversal isn't tested, I don't know
// a manga with enough cover images

var loadedTags = false

func tempRepo(t *testing.T, w io.Writer, ctx context.Context) Repository {
	t.Helper()
	return NewRepository(repositoryParams{ //nolint: contextcheck
		HttpClient: menou.DefaultClient,
		Cache:      mock.Cache{},
		Ctx:        ctx,
	}, zerolog.New(w))
}

func TestRepository_GetManga(t *testing.T) {
	r := tempRepo(t, io.Discard, t.Context())

	series, err := r.SeriesInfo(t.Context(), RainbowsAfterStormsID, payload.DownloadRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if series.Title != RainbowsAfterStorms {
		t.Errorf("got %s expected %s", series.Title, RainbowsAfterStorms)
	}
}

func TestRepository_GetCorrectStatus(t *testing.T) {
	r := tempRepo(t, io.Discard, t.Context())

	// My Intern Bullied Me Again!
	series, err := r.SeriesInfo(t.Context(), "c408ec80-586a-4d87-8bbc-e5e8d17a3898", payload.DownloadRequest{})
	if err != nil {
		t.Fatal(err)
	}

	want := float64(8)
	got, ok := series.HighestVolume.Get()
	if !ok {
		t.Errorf("highest volume is not set, should have been")
	}

	if want != got {
		t.Errorf("HighestVolume: got %f want %f", got, want)
	}

	want = float64(105)
	got, ok = series.HighestChapter.Get()
	if !ok {
		t.Errorf("highest chapter is not set, should have been")
	}

	if want != got {
		t.Errorf("HighestChapter: got %f want %f", got, want)
	}
}
