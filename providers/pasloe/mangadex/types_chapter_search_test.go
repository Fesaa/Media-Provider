package mangadex

import (
	"io"
	"testing"
)

func TestChapterSearchResponse_FilterOneEnChapter(t *testing.T) {
	r := tempRepo(t, io.Discard)

	res, err := r.GetChapters(RainbowsAfterStormsID)
	if err != nil {
		t.Fatal(err)
	}

	filtered := res.FilterOneEnChapter()
	if len(filtered.Data) != 172 {
		t.Errorf("Expected 172 chapters, got %d", len(filtered.Data))
	}
}

func TestChapterSearchResponse_FilterOneEnChapterSkipOfficial(t *testing.T) {

	c := ChapterSearchResponse{
		Result:   "",
		Response: "",
		Data: []ChapterSearchData{
			{
				Attributes: ChapterAttributes{
					ExternalUrl:        "some external url",
					TranslatedLanguage: "en",
				},
			},
		},
		Limit:  0,
		Offset: 0,
		Total:  0,
	}

	filtered := c.FilterOneEnChapter()
	if len(filtered.Data) != 0 {
		t.Errorf("Expected 0 chapters, got %d", len(filtered.Data))
	}

}
