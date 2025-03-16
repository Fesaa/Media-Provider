package webtoon

import (
	"github.com/rs/zerolog"
	"io"
	"net/http"
)

func tempRepository(w io.Writer) Repository {
	return NewRepository(http.DefaultClient, zerolog.New(w))
}

/*
func TestRepository_Search(t *testing.T) {
	repo := tempRepository(io.Discard)

	got, err := repo.Search(context.Background(), SearchOptions{Query: WebToonName})
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			t.Skipf("skipping due to rate limit")
		}
		t.Fatal(err)
	}

	if len(got) < 1 {
		t.Fatalf("got %d results, wanted at least one 1", len(got))
	}

	if got[0].Name != WebToonName {
		t.Errorf("got %q, want %q", got[0].Name, WebToonName)
	}
}

func TestRepository_LoadImages(t *testing.T) {
	repo := tempRepository(io.Discard)
	chpt := Chapter{
		Url:      "https://www.webtoons.com/en/romance/night-owls-and-summer-skies/episode-1/viewer?title_no=4747&episode_no=1",
		ImageUrl: "https://webtoon-phinf.pstatic.net/20220920_110/1663628756390wh0Oz_PNG/thumb_16636121077014747110.png?type=q90",
		Title:    "Episode 1",
		Number:   "1",
		Date:     "Sep 21, 2022",
	}

	time.Sleep(1 * time.Second)
	got, err := repo.LoadImages(context.Background(), chpt)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			t.Skipf("skipping due to rate limit")
		}
		t.Fatal(err)
	}

	want := 60
	if len(got) != want {
		t.Fatalf("got %d results, want %d", len(got), want)
	}
}

func TestRepository_SeriesInfoShort(t *testing.T) {
	repo := tempRepository(io.Discard)

	time.Sleep(1 * time.Second)
	got, err := repo.SeriesInfo(context.Background(), WebToonID)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			t.Skipf("skipping due to rate limit")
		}
		t.Fatal(err)
	}

	if got.Name != WebToonName {
		t.Errorf("got %q, want %q", got.Name, WebToonName)
	}

	want := "Despite her tough exterior, seventeen year-old Emma Lane has never been the outdoorsy type. So when her mother unceremoniously dumps her at Camp Mapplewood for the summer, she’s determined to get kicked out fast. However, when she draws the attention of Vivian Black, a mysterious and gorgeous assistant counselor, she discovers that there may be more to this camp than mean girls and mosquitos. There might even be love."
	if got.Description != want {
		t.Errorf("got %q, want %q", got.Description, want)
	}

	if len(got.Chapters) != 8 {
		t.Errorf("got %d results, want 8", len(got.Chapters))
	}

	if !got.Completed {
		t.Errorf("got %v, want true", got.Completed)
	}

	want = "Romance"
	if got.Genre != want {
		t.Errorf("got %q, want %q", got.Genre, want)
	}

	authors := []string{"TIKKLIL", "Rebecca Sullivan"}

	if !reflect.DeepEqual(got.Authors, authors) {
		t.Errorf("got %q, want %q", got.Authors, authors)
	}

}

func TestRepository_SeriesInfoHrefAuthor(t *testing.T) {
	repo := tempRepository(io.Discard)

	time.Sleep(1 * time.Second)
	got, err := repo.SeriesInfo(context.Background(), "6202")
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			t.Skipf("skipping due to rate limit")
		}
		t.Fatal(err)
	}

	if got.Name != "Osora" {
		t.Errorf("got %q, want %q", got.Name, "Osora")
	}

	if len(got.Chapters) < 48 {
		t.Errorf("got %d results, want at least 48", len(got.Chapters))
	}

	if !slices.Contains(got.Authors, "ToniRenea") {
		t.Errorf("got %q, want %q", got.Authors, "ToniRenea")
	}
}

func TestRepository_SeriesInfoSuperLong(t *testing.T) {
	repo := tempRepository(io.Discard)

	time.Sleep(1 * time.Second)
	got, err := repo.SeriesInfo(context.Background(), "4464")
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			t.Skipf("skipping due to rate limit")
		}
		t.Fatal(err)
	}

	if len(got.Chapters) < 114 {
		t.Errorf("got %d results, want at least 114", len(got.Chapters))
	}
}
*/
