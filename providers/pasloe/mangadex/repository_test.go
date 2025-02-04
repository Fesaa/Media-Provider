package mangadex

import (
	"context"
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
	t.Helper()
	return NewRepository(http.DefaultClient, zerolog.New(w))
}

func TestRepository_GetManga(t *testing.T) {
	r := tempRepo(t, io.Discard)

	res, err := r.GetManga(context.Background(), RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	if res.Data.Attributes.LangTitle("en") != RainbowsAfterStorms {
		t.Errorf("got %s expected %s", res.Data.Attributes.LangTitle("en"), RainbowsAfterStorms)
	}
}

func TestRepository_SearchManga(t *testing.T) {
	r := tempRepo(t, io.Discard)
	_, err := r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: true})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: false, IncludedTags: []string{"Romance"}})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: false, IncludedTags: []string{"FRYTGUHIJO"}})
	if err == nil {
		t.Fatal("expected error")
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: true, IncludedTags: []string{"FRYTGUHIJO"}})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: false, ExcludedTags: []string{"Romance"}})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: false, ExcludedTags: []string{"FRYTGUHIJO"}})
	if err == nil {
		t.Fatal("expected error")
	}

	time.Sleep(time.Second)
	_, err = r.SearchManga(context.Background(), SearchOptions{SkipNotFoundTags: true, ExcludedTags: []string{"FRYTGUHIJO"}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRepository_GetChapters(t *testing.T) {
	r := tempRepo(t, io.Discard)
	res, err := r.GetChapters(context.Background(), RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Data) != 372 {
		t.Errorf("got %d expected %d", len(res.Data), 372)
	}
}

func TestRepository_GetCoverImages(t *testing.T) {
	r := tempRepo(t, io.Discard)

	res, err := r.GetCoverImages(context.Background(), RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Data) != 13 {
		t.Errorf("got %d expected %d", len(res.Data), 13)
	}
}

func TestRepository_GetCoverImagesOffSet(t *testing.T) {
	r := tempRepo(t, io.Discard)
	res, err := r.GetCoverImages(context.Background(), RainbowsAfterStormsID, 13)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Data) != 0 {
		t.Errorf("got %d expected %d", len(res.Data), 0)
	}
}

func TestRepository_GetChapterImages(t *testing.T) {
	r := tempRepo(t, io.Discard)
	res, err := r.GetChapterImages(context.Background(), RainbowsAfterStormsLastChapterID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.FullImageUrls()) != 22 {
		t.Errorf("got %d expected %d", len(res.FullImageUrls()), 22)
	}
}
