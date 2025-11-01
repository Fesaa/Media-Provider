package mangadex

import (
	"context"
	"io"
	"testing"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/utils/mock"
	"github.com/rs/zerolog"
)

func tempRepo(t *testing.T, w io.Writer, ctx context.Context) Repository {
	t.Helper()
	return NewRepository(repositoryParams{ //nolint: contextcheck
		HttpClient: menou.DefaultClient,
		Cache:      mock.Cache{},
		Ctx:        ctx,
	}, zerolog.New(w))
}

func TestRepository_SeriesInfo_Writers(t *testing.T) {
	r := tempRepo(t, io.Discard, t.Context())

	series, err := r.SeriesInfo(t.Context(), "3adf74e1-ec07-4ad6-bdbb-0fdf6b4c5f53", payload.DownloadRequest{})
	if err != nil {
		t.Fatal(err)
	}

	lilyClub, ok := utils.FindOk(series.People, func(person publication.Person) bool {
		return person.Roles.HasRole(comicinfo.Writer)
	})

	if !ok {
		t.Fatal("could not find person")
	}

	if lilyClub.Name != "Lily Club (橘姬社)" {
		t.Errorf("lily club name: %s", lilyClub.Name)
	}

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
