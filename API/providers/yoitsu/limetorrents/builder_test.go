package limetorrents

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

func tempBuilder(w io.Writer) *Builder {
	return NewBuilder(zerolog.New(w), menou.DefaultClient, nil)
}

func TestBuilder_Search(t *testing.T) {
	b := tempBuilder(io.Discard)

	data, err := b.Search(SearchOptions{
		Category: ALL,
		Query:    "Modern Love S01",
		Page:     1,
	})
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("Skipping test, cannot reach")
		}
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("no results")
	}

	want := "Modern Love S01 COMPLETE 720p AMZN WEBRip x264-GalaxyTV[TGx]"
	got := utils.Find(data, func(result SearchResult) bool {
		return result.Name == want
	})

	if got == nil {
		t.Fatal("Can't find wanted result")
	}

	want = "28A3DF366D802741101E9E98C0B8CDB264B9DAA9"
	if got.Hash != want {
		t.Errorf("got %s, want %s", got.Hash, want)
	}

	want = "1.62 GB"
	if got.Size != want {
		t.Errorf("got %s, want %s", got.Size, want)
	}
}

func TestBuilder_Transform(t *testing.T) {
	b := tempBuilder(io.Discard)

	got := b.Transform(payload.SearchRequest{
		Provider:  []models.Provider{models.LIME},
		Query:     "Modern Love S01",
		Modifiers: nil,
	})

	want := SearchOptions{
		Category: ALL,
		Query:    "Modern Love S01",
		Page:     1,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	got = b.Transform(payload.SearchRequest{
		Provider: []models.Provider{models.LIME},
		Query:    "Modern Love S01",
		Modifiers: map[string][]string{
			"categories": {MOVIE},
		},
	})

	want = SearchOptions{
		Category: MOVIE,
		Query:    "Modern Love S01",
		Page:     1,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

}
