package webtoon

import (
	"reflect"
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
)

const (
	WebToonID   = "4747"
	WebToonName = "Night Owls & Summer Skies"
)

func TestBuilder_Provider(t *testing.T) {
	b := &Builder{}

	got := b.Provider()
	want := models.WEBTOON
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuilder_Normalize(t *testing.T) {
	b := &Builder{}

	data := []SearchData{
		{
			Id:              WebToonID,
			Name:            WebToonName,
			ReadCount:       "",
			ThumbnailMobile: "",
			AuthorNameList:  []string{"TIKKLIL", "Rebecca Sullivan"},
			Genre:           "Romance",
			Rating:          false,
		},
	}

	got := b.Normalize(t.Context(), data)

	if len(got) != len(data) {
		t.Errorf("got %d results, want %d", len(got), len(data))
	}

	first := got[0]

	if first.Name != WebToonName {
		t.Errorf("got %q, want %q", first.Name, WebToonName)
	}

	if utils.Find(first.Tags, func(tag payload.InfoTag) bool {
		return tag.Name == "Genre"
	}) == nil {
		t.Errorf("got %q, need to include genre tag", first.Tags)
	}

	if first.InfoHash != WebToonID {
		t.Errorf("got %q, want %q", first.InfoHash, WebToonID)
	}
}

func TestBuilder_Transform(t *testing.T) {
	b := &Builder{}

	want := SearchOptions{Query: WebToonName}
	got := b.Transform(t.Context(), payload.SearchRequest{
		Provider:  []models.Provider{models.WEBTOON},
		Query:     WebToonName,
		Modifiers: nil,
	})

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
