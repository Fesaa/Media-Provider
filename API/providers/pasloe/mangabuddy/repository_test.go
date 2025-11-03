package mangabuddy

import (
	"io"
	"strings"
	"testing"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
)

func tempRepository(w io.Writer) Repository {
	log := zerolog.New(w)
	return NewRepository(menou.DefaultClient, log, services.MarkdownServiceProvider(log))
}

func Test_repository_Search(t *testing.T) {
	repo := tempRepository(io.Discard)

	series, err := repo.Search(t.Context(), SearchOptions{Query: "baili jin"})
	if err != nil {
		t.Fatal(err)
	}

	if len(series) == 0 {
		t.Fatal("no series")
	}

	firstSeries := series[0]

	want := "Baili Jin Among Mortals"
	got := firstSeries.Name
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	want = "Baili Jin, a fairy who was living in heaven, eating and drinking without a care, broke her Majesty's colourful, stained-glass plate at her birthday and got "
	got = firstSeries.Description
	if !strings.HasPrefix(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}

	info, err := repo.SeriesInfo(t.Context(), firstSeries.InfoHash, payload.DownloadRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if info.People[0].Name != "Julys" {
		t.Errorf("got %q, want Julys", info.People[0].Name)
	}

	imgs, err := repo.ChapterUrls(t.Context(), publication.Chapter{
		Id: "/baili-jin-among-mortals/chapter-213",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(imgs) != 38 {
		t.Errorf("got %d, want 38", len(imgs))
	}

}
