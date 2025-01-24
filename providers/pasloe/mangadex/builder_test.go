package mangadex

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"reflect"
	"testing"
)

const (
	RainbowsAfterStorms              = "Rainbows After Storms"
	RainbowsAfterStormsID            = "bc86a871-ddc5-4e42-812a-ccd38101d82e"
	RainbowsAfterStormsLastChapterID = "7d327897-5903-4fa1-92d7-f01c3c686a36"
)

func TestBuilder_Transform(t *testing.T) {
	tests := []struct {
		name string
		args payload.SearchRequest
		want SearchOptions
	}{
		{
			name: "TestQuery",
			args: payload.SearchRequest{
				Query: "MyQuery",
			},
			want: SearchOptions{
				Query:            "MyQuery",
				SkipNotFoundTags: true,
			},
		},
		{
			name: "TestSkipNotFoundTrue",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"SkipNotFoundTags": {"true"},
				},
			},
			want: SearchOptions{
				SkipNotFoundTags: true,
			},
		},
		{
			name: "TestSkipNotFoundFalse",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"SkipNotFoundTags": {"false"},
				},
			},
			want: SearchOptions{
				SkipNotFoundTags: false,
			},
		},
		{
			name: "TestSkipNotFoundDefault",
			args: payload.SearchRequest{},
			want: SearchOptions{
				SkipNotFoundTags: true,
			},
		},
		{
			name: "TestIncludeTags",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"includeTags": {"tag1", "tag2"},
				},
			},
			want: SearchOptions{SkipNotFoundTags: true, IncludedTags: []string{"tag1", "tag2"}},
		},
		{
			name: "TestExcludeTags",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"excludeTags": {"tag1", "tag2"},
				},
			},
			want: SearchOptions{SkipNotFoundTags: true, ExcludedTags: []string{"tag1", "tag2"}},
		},
		{
			name: "TestStatus",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"status": {"Completed"},
				},
			},
			want: SearchOptions{SkipNotFoundTags: true, Status: []string{"Completed"}},
		},
		{
			name: "TestContentRating",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"contentRating": {"safe"},
				},
			},
			want: SearchOptions{SkipNotFoundTags: true, ContentRating: []string{"safe"}},
		},
		{
			name: "TestDemographic",
			args: payload.SearchRequest{
				Modifiers: map[string][]string{
					"publicationDemographic": {"josei"},
				},
			},
			want: SearchOptions{SkipNotFoundTags: true, PublicationDemographic: []string{"josei"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{}
			if got := b.Transform(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Transform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_NormalizeNil(t *testing.T) {
	b := &Builder{}

	got := b.Normalize(nil)
	if got == nil {
		t.Fatal("Normalize(nil) returned nil")
	}

	if len(got) != 0 {
		t.Fatal("Normalize(nil) returned non-zero length")
	}
}

func TestBuilder_Normalize(t *testing.T) {
	b := &Builder{}

	repo := NewRepository(http.DefaultClient, zerolog.New(io.Discard))

	res, err := repo.SearchManga(SearchOptions{Query: RainbowsAfterStorms})
	if err != nil {
		t.Fatal(err)
	}

	got := b.Normalize(res)

	if len(got) != len(res.Data) {
		t.Fatalf("Normalize(%+v): got %+v, expected %+v", res, got, res.Data)
	}

	bestManga := got[0]

	if bestManga.Name != RainbowsAfterStorms {
		t.Errorf("got %s, expected %s", bestManga.Name, RainbowsAfterStorms)
	}

	want := "Unbeknownst to their friends and classmates, Nanoha and Chidori are dating. Watch as they try to keep their relationship secret!"
	if bestManga.Description != want {
		t.Errorf("got %s, expected %s", bestManga.Description, want)
	}

	want = "13 Vol. 162 Ch."
	if bestManga.Size != want {
		t.Errorf("got %s, expected %s", bestManga.Size, want)
	}

	if bestManga.InfoHash != RainbowsAfterStormsID {
		t.Errorf("got %s, expected %s", bestManga.InfoHash, RainbowsAfterStormsID)
	}

	want = "https://mangadex.org/title/bc86a871-ddc5-4e42-812a-ccd38101d82e/"
	if bestManga.RefUrl != want {
		t.Errorf("got %s, expected %s", bestManga.RefUrl, want)
	}

}

func TestBuilder_Provider(t *testing.T) {
	b := &Builder{}
	if b.Provider() != models.MANGADEX {
		t.Errorf("got %s, expected %s", b.Provider(), models.MANGADEX)
	}
}
