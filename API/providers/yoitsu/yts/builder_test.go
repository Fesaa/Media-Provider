package yts

import (
	"reflect"
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
)

func TestBuilder_Transform(t *testing.T) {
	want := SearchOptions{
		Query:  "We live in time",
		SortBy: "title",
		Page:   1,
	}

	got := (&Builder{}).Transform(payload.SearchRequest{
		Provider: []models.Provider{models.YTS},
		Query:    "We live in time",
		Modifiers: map[string][]string{
			"sortBys": {"title", "FGHJK"},
		},
	})

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v\nwant: %v", got, want)
	}
}
