package bato

import (
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"strings"
	"testing"
)

func tempRepository(w io.Writer) Repository {
	return NewRepository(http.DefaultClient, zerolog.New(w))
}

func TestRepository_Search(t *testing.T) {
	repo := tempRepository(io.Discard)

	options := SearchOptions{
		Query:              "Heart",
		Genres:             []string{"yuri"},
		OriginalLang:       []string{"zh"},
		TranslatedLang:     []string{"en"},
		OriginalWorkStatus: []Publication{PublicationOngoing},
		BatoUploadStatus:   []Publication{PublicationOngoing},
	}

	series, err := repo.Search(t.Context(), options)
	if err != nil {
		t.Fatal(err)
	}

	if len(series) < 3 {
		t.Errorf("got %d series, want at least 3", len(series))
	}

	heartOfThorns := utils.Find(series, func(result SearchResult) bool {
		return result.Title == "Heart Of Thorns"
	})

	if heartOfThorns == nil {
		t.Fatalf("got no heart of thorns")
	}

	if heartOfThorns.Id != "172024-heart-of-thorns" {
		t.Errorf("wanted %s got %s", "172024-heart-of-thorns", heartOfThorns.Id)
	}

}

func TestRepository_SeriesInfo(t *testing.T) {
	repo := tempRepository(io.Discard)

	series, err := repo.SeriesInfo(t.Context(), "172024-heart-of-thorns")
	if err != nil {
		t.Fatal(err)
	}

	if series.Id != "172024-heart-of-thorns" {
		t.Fatalf("wrong series id %s", series.Id)
	}

	if len(series.Chapters) < 70 {
		t.Fatalf("got %d series, want at least 70", len(series.Chapters))
	}

	if series.Title != "Heart Of Thorns" {
		t.Fatalf("got %s want %s", series.Title, "Heart Of Thorns")
	}

	if len(series.Tags) < 10 {
		t.Fatalf("got %d series, want at least 10", len(series.Tags))
	}

	comma := utils.Find(series.Tags, func(s string) bool {
		return s == ","
	})

	if comma != nil {
		t.Fatalf("got %s want nil", *comma)
	}

	if series.Chapters[0].Volume != "" {
		t.Fatalf("got %s want nil", series.Chapters[0].Volume)
	}

	lilyClub := utils.Find(series.Authors, func(s Author) bool {
		return s.Name == "Lily Club (橘姬社)"
	})

	if lilyClub == nil {
		t.Fatalf("got no Lily Club")
	}

	cleaned := utils.Find(series.Authors, func(s Author) bool {
		return s.Name == "狐泥"
	})

	if cleaned == nil {
		t.Fatalf("got no cleaned Authors")
	}

	notCleaned := utils.Find(series.Authors, func(s Author) bool {
		return strings.Contains(s.Name, "(Story&Art)")
	})

	if notCleaned != nil {
		t.Fatalf("got %s want nil", *notCleaned)
	}

}

func TestRepository_SeriesWithVolume(t *testing.T) {
	repo := tempRepository(io.Discard)

	series, err := repo.SeriesInfo(t.Context(), "133980-little-mushroom")
	if err != nil {
		t.Fatal(err)
	}

	if series.Id != "133980-little-mushroom" {
		t.Fatalf("got %s want %s", series.Id, "133980-little-mushroom")
	}

	if len(series.Chapters) < 10 {
		t.Fatalf("got %d series, want at least 10", len(series.Chapters))
	}

	if series.Chapters[0].Volume != "1" {
		t.Fatalf("got %s wanted 1", series.Chapters[0].Volume)
	}
}

func TestRepository_ChapterImages(t *testing.T) {
	repo := tempRepository(io.Discard)

	images, err := repo.ChapterImages(t.Context(), "172024-heart-of-thorns/3343778-ch_77")
	if err != nil {
		t.Fatal(err)
	}

	if len(images) == 0 {
		t.Fatalf("got no images")
	}

}
