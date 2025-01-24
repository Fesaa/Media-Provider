package mangadex

import (
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"testing"
	"time"
)

// Note: Sleeping quite a bit during test, to ensure we do not send
// too many requests to MangeDex
// Do NOT run these in parallel
//
// Repository.GetCoverImages's offset traversal isn't tested, I don't know
// a manga with enough cover images

var loadedTags = false

func tempRepo(t *testing.T, w io.Writer) Repository {
	if !loadedTags {
		if err := loadTags(http.DefaultClient); err != nil {
			t.Fatal(err)
		}
		loadedTags = true
	}
	return NewRepository(http.DefaultClient, zerolog.New(w))
}

func TestRepository_GetManga(t *testing.T) {
	r := tempRepo(t, io.Discard)

	res, err := r.GetManga(RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	if res.Data.Attributes.EnTitle() != RainbowsAfterStorms {
		t.Errorf("got %s expected %s", res.Data.Attributes.EnTitle(), RainbowsAfterStorms)
	}
}

func TestRepository_SearchManga(t *testing.T) {
	r := tempRepo(t, io.Discard)
	_, err := r.SearchManga(SearchOptions{SkipNotFoundTags: true})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(SearchOptions{SkipNotFoundTags: false, IncludedTags: []string{"Romance"}})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(SearchOptions{SkipNotFoundTags: false, IncludedTags: []string{"FRYTGUHIJO"}})
	if err == nil {
		t.Fatal("expected error")
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(SearchOptions{SkipNotFoundTags: true, IncludedTags: []string{"FRYTGUHIJO"}})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(SearchOptions{SkipNotFoundTags: false, ExcludedTags: []string{"Romance"}})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(SearchOptions{SkipNotFoundTags: false, ExcludedTags: []string{"FRYTGUHIJO"}})
	if err == nil {
		t.Fatal("expected error")
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(SearchOptions{SkipNotFoundTags: true, ExcludedTags: []string{"FRYTGUHIJO"}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRepository_GetChapters(t *testing.T) {
	r := tempRepo(t, io.Discard)
	res, err := r.GetChapters(RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Data) != 372 {
		t.Errorf("got %d expected %d", len(res.Data), 372)
	}
}

func TestRepository_GetCoverImages(t *testing.T) {
	r := tempRepo(t, io.Discard)

	res, err := r.GetCoverImages(RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Data) != 13 {
		t.Errorf("got %d expected %d", len(res.Data), 13)
	}
}

func TestRepository_GetCoverImagesOffSet(t *testing.T) {
	r := tempRepo(t, io.Discard)
	res, err := r.GetCoverImages(RainbowsAfterStormsID, 13)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Data) != 0 {
		t.Errorf("got %d expected %d", len(res.Data), 0)
	}
}

func TestRepository_GetChapterImages(t *testing.T) {
	r := tempRepo(t, io.Discard)
	res, err := r.GetChapterImages(RainbowsAfterStormsLastChapterID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.FullImageUrls()) != 22 {
		t.Errorf("got %d expected %d", len(res.FullImageUrls()), 22)
	}
}
