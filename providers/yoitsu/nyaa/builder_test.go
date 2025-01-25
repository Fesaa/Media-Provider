package nyaa

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/irevenko/go-nyaa/nyaa"
	"reflect"
	"testing"
)

func TestBuilder_Transform(t *testing.T) {
	want := nyaa.SearchOptions{
		Provider: "nyaa",
		Query:    "Spice+and+Wolf",
		Category: "literature-eng",
		SortBy:   "downloads",
		Filter:   "no-filter",
	}

	got := (&Builder{}).Transform(payload.SearchRequest{
		Provider: []models.Provider{models.NYAA},
		Query:    "Spice and Wolf",
		Modifiers: map[string][]string{
			"categories": {"literature-eng", "DFGHJK"},
			"sortBys":    {"downloads"},
			"filters":    {"no-filter"},
		},
	})

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Transform() = %+v, want %+v", got, want)
	}
}
