package dynasty

import (
	"bytes"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

const (
	// For tests with special non-chapter; https://dynasty-scans.com/series/shiawase_trimming
	ShiawaseTrimming   = "Shiawase Trimming"
	ShiawaseTrimmingId = "/series/shiawase_trimming"
)

func tempRepository(w io.Writer) Repository {
	return NewRepository(menou.DefaultClient, zerolog.New(w))
}

func TestRepository_SearchSeries(t *testing.T) {
	time.Sleep(1 * time.Second)
	var buf bytes.Buffer
	repo := tempRepository(&buf)

	series, err := repo.SearchSeries(t.Context(), SearchOptions{Query: "Sailor Girlfriend"})
	if err != nil {
		t.Fatalf("SearchSeries: %v", err)
	}

	if len(series) != 1 {
		t.Fatalf("SearchSeries: expected 1 series, got %d", len(series))
	}

	serie := series[0]

	if utils.Find(serie.Tags, func(tag publication.Tag) bool {
		return tag.Value == "Yuri"
	}) == nil {
		t.Fatalf("SearchSeries: expected tag with display name 'Yuri', got nil")
	}

	if utils.Find(serie.Authors, func(author publication.Person) bool {
		return author.Name == "Kanbayashi Makoto"
	}) == nil {
		t.Fatalf("SearchSeries: expected author with display name 'Kanbayashi Makoto', got nil")
	}
}

func TestRepository_SeriesInfo(t *testing.T) {
	time.Sleep(1 * time.Second)
	var buf bytes.Buffer
	repo := tempRepository(&buf)

	info, err := repo.SeriesInfo(t.Context(), "/series/sailor_girlfriend", payload.DownloadRequest{})
	if err != nil {
		t.Fatalf("SeriesInfo: %v", err)
	}

	if len(info.Chapters) != 5 {
		t.Fatalf("SeriesInfo: expected 5 chapters, got %d", len(info.Chapters))
	}

	if info.Status != publication.StatusCompleted {
		t.Fatalf("SeriesInfo: expected Completed status, got %s", info.Status)
	}

	firstChapter := info.Chapters[0]
	if firstChapter.ReleaseDate == nil {
		t.Fatalf("SeriesInfo: expected release date, got nil")
	}
	firstChapterReleaseDate := utils.MustReturn(time.Parse(RELEASEDATAFORMAT, "May 29 '18"))
	if *firstChapter.ReleaseDate != firstChapterReleaseDate {
		t.Fatalf("SeriesInfo: expected release date (May 29 '18), got %s", firstChapter.ReleaseDate)
	}

	lastChapter := info.Chapters[4]
	if lastChapter.ReleaseDate == nil {
		t.Fatalf("SeriesInfo: expected release date, got nil")
	}

	lastChapterReleaseDate := utils.MustReturn(time.Parse(RELEASEDATAFORMAT, "Jan 22 '25"))
	if *lastChapter.ReleaseDate != lastChapterReleaseDate {
		t.Fatalf("SeriesInfo: expected release date (Jan 22 '25), got %s", lastChapter.ReleaseDate)
	}

	want := "A Sailor's Girlfriend's Day"
	if lastChapter.Title != want {
		t.Fatalf("SeriesInfo: expected %s got %s", want, lastChapter.Title)
	}

	if utils.Find(lastChapter.Tags, func(tag publication.Tag) bool {
		return tag.Value == "Drunk"
	}) == nil {
		t.Fatalf("SeriesInfo: Expected last chapter to have tag Drunk got %+v", lastChapter.Tags)
	}

}

func TestRepository_SeriesInfoWithVolumes(t *testing.T) {
	time.Sleep(1 * time.Second)
	var buf bytes.Buffer
	repo := tempRepository(&buf)

	info, err := repo.SeriesInfo(t.Context(), "/series/canaries_dream_of_shining_stars", payload.DownloadRequest{})
	if err != nil {
		t.Fatalf("SeriesInfo: %v", err)
	}

	if len(info.Chapters) < 8 {
		t.Fatalf("SeriesInfo: expected at least 8 chapters, got %d", len(info.Chapters))
	}

	firstChapter := info.Chapters[0]

	if firstChapter.Volume != "1" {
		t.Fatalf("SeriesInfo: expected volume to be 1, got %s", firstChapter.Volume)
	}

	secondVolumeChapter := info.Chapters[6]
	if secondVolumeChapter.Volume != "2" {
		t.Fatalf("SeriesInfo: expected volume to be 2, got %s", secondVolumeChapter.Volume)
	}

}

func TestRepository_ChapterImages(t *testing.T) {
	time.Sleep(1 * time.Second)
	var buf bytes.Buffer
	repo := tempRepository(&buf)

	images, err := repo.ChapterUrls(t.Context(), publication.Chapter{
		Id: "canaries_dream_of_shining_stars_ch01",
	})
	if err != nil {
		if strings.Contains(err.Error(), "status code error: 503") {
			t.Skipf("Skipping test as 3rd party server error")
		}

		t.Fatalf("ChapterImages: %v", err)
	}

	if len(images) != 43 {
		t.Fatalf("ChapterImages: expected 43 images, got %d", len(images))
	}
}

func TestRepository_SearchSeriesOneShotChapters(t *testing.T) {
	time.Sleep(1 * time.Second)
	var buf bytes.Buffer
	repo := tempRepository(&buf)

	res, err := repo.SeriesInfo(t.Context(), ShiawaseTrimmingId, payload.DownloadRequest{})
	if err != nil {
		if strings.Contains(err.Error(), "status code error: 503") {
			t.Skipf("Skipping test as 3rd party server error")
		}
		t.Fatalf("SeriesInfo: %v", err)
	}

	if len(res.Chapters) < 23 {
		t.Fatalf("SeriesInfo: expected at least 23 chapters, got %d", len(res.Chapters))
	}

	firstChapter := res.Chapters[0]

	want := "Manga Time Kirara 20th Anniversary Special Collaboration: Stardust Telepath x Shiawase Trimming"
	if firstChapter.Title != want {
		t.Errorf("SeriesInfo: expected %s got %s", want, firstChapter.Title)
	}

	if firstChapter.Volume != "" || firstChapter.Chapter != "" {
		t.Errorf("SeriesInfo: expected empty chapter got Vol. %s Ch. %s", firstChapter.Volume, firstChapter.Chapter)
	}

	if len(firstChapter.People) != 2 {
		t.Errorf("SeriesInfo: expected 2 authors, got %d", len(firstChapter.People))
	}

	if len(firstChapter.Tags) != 1 {
		t.Errorf("SeriesInfo: expected 1 tags, got %d", len(firstChapter.Tags))
	}

}

func TestRepository_ChapterInfo(t *testing.T) {
	time.Sleep(1 * time.Second)
	var buf bytes.Buffer
	repo := tempRepository(&buf)

	s, err := repo.SeriesInfo(t.Context(), "/chapters/assorted_yujin_x_tenma_drawings_p_o_p_o", payload.DownloadRequest{})
	if err != nil {
		if strings.Contains(err.Error(), "status code error: 503") {
			t.Skipf("Skipping test as 3rd party server error")
		}
		t.Fatalf("SeriesInfo: %v", err)
	}

	want := "Assorted Yujin x Tenma Drawings (p o p o)"
	got := s.Title
	if want != got {
		t.Errorf("SeriesInfo: expected %s got %s", want, got)
	}

	if len(s.Chapters) != 1 {
		t.Errorf("SeriesInfo: expected 1 chapter, got %d", len(s.Chapters))
	}

	want = "Me-A Scans"
	got = s.Chapters[0].Translator[0]
	if want != got {
		t.Errorf("SeriesInfo: expected %s got %s", want, got)
	}

	tags := utils.Map(s.Tags, func(t publication.Tag) string {
		return t.Value
	})

	if !slices.Contains(tags, "Height gap") {
		t.Errorf("SeriesInfo: expected %s tags to contain height gap", tags)
	}

	if s.Year != 2021 {
		t.Errorf("SeriesInfo: expected 2021, got %d", s.Year)
	}
}
