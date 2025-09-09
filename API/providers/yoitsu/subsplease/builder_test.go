package subsplease

import (
	"testing"

	"github.com/Fesaa/Media-Provider/http/menou"
)

// Subsplease does weird stuff when not having any results.
func TestBuilder_SearchEmpty(t *testing.T) {
	got, err := (&Builder{httpClient: menou.DefaultClient}).Search(t.Context(), SearchOptions{
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
	got, err := (&Builder{httpClient: menou.DefaultClient}).Search(t.Context(), SearchOptions{
		Query: "Spice and Wolf",
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(got) == 0 {
		t.Fatalf("got %d results, want at least 1 result", len(got))
	}
}
