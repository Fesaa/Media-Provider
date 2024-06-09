package yts

import (
	"net/url"
	"testing"
)

func TestReq(t *testing.T) {
	opt := YTSSearchOptions{
		Query:  url.QueryEscape("Star Wars"),
		SortBy: "Seeders",
		Page:   0,
	}
	search, err := Search(opt)
	if err != nil {
		t.Fatal(err)
	}
	if search.Data.MovieCount == 0 {
		t.Fatal("No movies found")
	}
	found := false
	for _, movie := range search.Data.Movies {
		if movie.Title == "Rogue One: A Star Wars Story" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Unable to find: Rogue One: A Star Wars Story")
	}
}
