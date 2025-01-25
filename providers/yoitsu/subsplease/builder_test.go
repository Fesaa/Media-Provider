package subsplease

import (
	"net/http"
	"testing"
)

// Subsplease does weird stuff when not having any results.
func TestBuilder_SearchEmpty(t *testing.T) {
	got, err := (&Builder{httpClient: http.DefaultClient}).Search(SearchOptions{
		Query: "Something they don't have",
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 0 {
		t.Fatalf("got %d results, want 0", len(got))
	}
}

func TestBuilder_Search(t *testing.T) {
	got, err := (&Builder{httpClient: http.DefaultClient}).Search(SearchOptions{
		Query: "Spice and Wolf",
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(got) == 0 {
		t.Fatalf("got %d results, want at least 1 result", len(got))
	}
}
