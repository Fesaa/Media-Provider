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

func getChapter(series *Series, chapter string, volume ...string) *Chapter {
	return utils.Find(series.Chapters, func(c Chapter) bool {
		if len(volume) > 0 {
			return c.Volume == volume[0] && c.Chapter == chapter
		}

		return c.Chapter == chapter
	})
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

func TestRepository_SeriesInfo_ChapterDecimals(t *testing.T) {
	repo := tempRepository(io.Discard)
	series, err := repo.SeriesInfo(t.Context(), "99749-still-sick-official")
	if err != nil {
		t.Fatal(err)
	}

	want := 26
	if len(series.Chapters) != want {
		t.Fatalf("got %d series, want %d", len(series.Chapters), want)
	}

	groupByChapter := utils.GroupBy(series.Chapters, func(chapter Chapter) string {
		return chapter.Chapter
	})

	for _, chapter := range groupByChapter {
		if len(chapter) != 1 {
			t.Fatalf("got a duplicate chapter %s: %d", chapter[0].Chapter, len(chapter))
		}
	}

	volume3Chapter23Point5 := getChapter(series, "23.5", "3")

	if volume3Chapter23Point5 == nil {
		t.Fatalf("failed to find volume 3 chapter 23.5")
	}

}

func TestRepository_SeriesInfoFullyAnchoredTitles(t *testing.T) {
	repo := tempRepository(io.Discard)

	series, err := repo.SeriesInfo(t.Context(), "149409-romance-of-the-stars-official")
	if err != nil {
		t.Fatal(err)
	}

	chapter12 := getChapter(series, "12")

	if chapter12 == nil {
		t.Fatalf("failed to find chapter 12")
	}

	want := "Yitong, I'm Sorry, But"
	got := chapter12.Title

	if got != want {
		t.Fatalf("got \"%s\" want \"%s\"", got, want)
	}

	chapter5 := getChapter(series, "5")
	if chapter5 == nil {
		t.Fatalf("failed to find chapter 5")
	}

	want = "Jealous Friend"
	got = chapter5.Title
	if got != want {
		t.Fatalf("got \"%s\" want \"%s\"", got, want)
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

func TestRepository_SeriesInfoEpisodesAndSeasons(t *testing.T) {
	repo := tempRepository(io.Discard)

	series, err := repo.SeriesInfo(t.Context(), "148408-mage-demon-queen-official")
	if err != nil {
		t.Fatal(err)
	}

	s1E12 := getChapter(series, "12", "1")
	if s1E12 == nil {
		t.Fatalf("failed to find s1E12")
	}

	want := "Melathia Fanfic 1"
	got := s1E12.Title
	if got != want {
		t.Fatalf("got \"%s\" want \"%s\"", got, want)
	}

	s3E1 := getChapter(series, "1", "3")
	if s3E1 == nil {
		t.Fatalf("failed to find s3E1")
	}

	want = "Season 3 Premiere"
	got = s3E1.Title
	if got != want {
		t.Fatalf("got \"%s\" want \"%s\"", got, want)
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
