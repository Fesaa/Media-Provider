package mangadex

import (
	"testing"
)

func TestMangaSearchData_RefURL(t *testing.T) {
	manga := &MangaSearchData{Id: "test-id"}
	expectedURL := "https://mangadex.org/title/test-id/"
	if manga.RefURL() != expectedURL {
		t.Errorf("Expected RefURL to be '%s', got '%s'", expectedURL, manga.RefURL())
	}
}

func TestMangaSearchData_CoverURL(t *testing.T) {
	tests := []struct {
		name          string
		manga         *MangaSearchData
		expectedCover string
	}{
		{
			name: "Cover found",
			manga: &MangaSearchData{
				Id: "test-id",
				Relationships: []Relationship{
					{
						Type: "cover_art",
						Attributes: map[string]interface{}{
							"fileName": "cover.jpg",
						},
					},
				},
			},
			expectedCover: "proxy/mangadex/covers/test-id/cover.jpg.256.jpg",
		},
		{
			name: "No cover found",
			manga: &MangaSearchData{
				Id:            "test-id",
				Relationships: []Relationship{},
			},
			expectedCover: "",
		},
		{
			name: "Cover file name not string",
			manga: &MangaSearchData{
				Id: "test-id",
				Relationships: []Relationship{
					{
						Type: "cover_art",
						Attributes: map[string]interface{}{
							"fileName": 123,
						},
					},
				},
			},
			expectedCover: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.manga.CoverURL() != tt.expectedCover {
				t.Errorf("Expected CoverURL to be '%s', got '%s'", tt.expectedCover, tt.manga.CoverURL())
			}
		})
	}
}

func TestMangaSearchData_Authors(t *testing.T) {
	manga := &MangaSearchData{
		Relationships: []Relationship{
			{
				Type: "author",
				Attributes: map[string]interface{}{
					"name": "Author 1",
				},
			},
			{
				Type: "artist",
				Attributes: map[string]interface{}{
					"name": "Artist 1",
				},
			},
			{
				Type: "author",
				Attributes: map[string]interface{}{
					"name": "Author 2",
				},
			},
		},
	}

	expectedAuthors := []string{"Author 1", "Author 2"}
	authors := manga.Authors()

	if len(authors) != len(expectedAuthors) {
		t.Errorf("Expected %d authors, got %d", len(expectedAuthors), len(authors))
	}

	for i, author := range authors {
		if author != expectedAuthors[i] {
			t.Errorf("Expected author '%s', got '%s'", expectedAuthors[i], author)
		}
	}
}

func TestMangaSearchData_Artists(t *testing.T) {
	manga := &MangaSearchData{
		Relationships: []Relationship{
			{
				Type: "author",
				Attributes: map[string]interface{}{
					"name": "Author 1",
				},
			},
			{
				Type: "artist",
				Attributes: map[string]interface{}{
					"name": "Artist 1",
				},
			},
			{
				Type: "artist",
				Attributes: map[string]interface{}{
					"name": "Artist 2",
				},
			},
		},
	}

	expectedArtists := []string{"Artist 1", "Artist 2"}
	artists := manga.Artists()

	if len(artists) != len(expectedArtists) {
		t.Errorf("Expected %d artists, got %d", len(expectedArtists), len(artists))
	}

	for i, artist := range artists {
		if artist != expectedArtists[i] {
			t.Errorf("Expected artist '%s', got '%s'", expectedArtists[i], artist)
		}
	}
}

func TestMangaSearchData_ScanlationGroup(t *testing.T) {
	manga := &MangaSearchData{
		Relationships: []Relationship{
			{
				Type: "scanlation_group",
				Attributes: map[string]interface{}{
					"name": "Group 1",
				},
			},
			{
				Type: "artist",
				Attributes: map[string]interface{}{
					"name": "Artist 1",
				},
			},
			{
				Type: "scanlation_group",
				Attributes: map[string]interface{}{
					"name": "Group 2",
				},
			},
		},
	}

	expectedGroups := []string{"Group 1", "Group 2"}
	groups := manga.ScanlationGroup()

	if len(groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(groups))
	}

	for i, group := range groups {
		if group != expectedGroups[i] {
			t.Errorf("Expected group '%s', got '%s'", expectedGroups[i], group)
		}
	}
}

func TestMangaAttributes_LangTitle(t *testing.T) {
	attributes := MangaAttributes{
		Title: map[string]string{
			"en": "English Title",
			"ja": "Japanese Title",
		},
		AltTitles: []map[string]string{
			{"en": "English Alt Title"},
			{"ja": "Japanese Alt Title"},
		},
	}

	got := attributes.LangTitle("en")
	if got != "English Title" {
		t.Errorf("Expected English title, got '%s'", got)
	}

	attributes.Title["en"] = ""
	got = attributes.LangTitle("en")
	if got != "English Alt Title" {
		t.Errorf("Expected English alt title, got '%s'", got)
	}

	attributes.AltTitles = []map[string]string{}
	got = attributes.LangTitle("en")
	if got != "Japanese Title" {
		t.Errorf("Expected Japanese title, got '%s'", got)
	}

	attributes.Title = map[string]string{}
	got = attributes.LangTitle("en")
	if got != "Media-Provider-Fallback-title" {
		t.Errorf("Expected fallback title, got '%s'", got)
	}
}

func TestMangaAttributes_LangAltTitles(t *testing.T) {
	attributes := MangaAttributes{
		AltTitles: []map[string]string{
			{"en": "English Alt Title 1"},
			{"ja": "Japanese Alt Title 1"},
			{"en": "English Alt Title 2"},
		},
	}

	expectedAltTitles := []string{"English Alt Title 1", "English Alt Title 2"}
	altTitles := attributes.LangAltTitles("en")

	if len(altTitles) != len(expectedAltTitles) {
		t.Errorf("Expected %d alt titles, got %d", len(expectedAltTitles), len(altTitles))
	}

	for i, altTitle := range altTitles {
		if altTitle != expectedAltTitles[i] {
			t.Errorf("Expected alt title '%s', got '%s'", expectedAltTitles[i], altTitle)
		}
	}
}

func TestMangaAttributes_LangDescription(t *testing.T) {
	attributes := MangaAttributes{
		Description: map[string]string{
			"en": "English Description",
			"ja": "Japanese Description",
		},
	}

	if attributes.LangDescription("en") != "English Description" {
		t.Errorf("Expected English description, got '%s'", attributes.LangDescription("en"))
	}

	if attributes.LangDescription("fr") != "" {
		t.Errorf("Expected empty description, got '%s'", attributes.LangDescription("fr"))
	}
}

func TestContentRating_ComicInfoAgeRating(t *testing.T) {
	tests := []struct {
		name     string
		rating   ContentRating
		expected string
	}{
		{
			name:     "Safe",
			rating:   ContentRatingSafe,
			expected: "Everyone",
		},
		{
			name:     "Suggestive",
			rating:   ContentRatingSuggestive,
			expected: "Teen",
		},
		{
			name:     "Erotica",
			rating:   ContentRatingErotica,
			expected: "Mature 17+",
		},
		{
			name:     "Pornographic",
			rating:   ContentRatingPornographic,
			expected: "Adults Only 18+",
		},
		{
			name:     "Unknown",
			rating:   "unknown",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.rating.ComicInfoAgeRating())
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestContentRating_MetronInfoAgeRating(t *testing.T) {
	tests := []struct {
		name     string
		rating   ContentRating
		expected string
	}{
		{
			name:     "Safe",
			rating:   ContentRatingSafe,
			expected: "Everyone",
		},
		{
			name:     "Suggestive",
			rating:   ContentRatingSuggestive,
			expected: "Mature",
		},
		{
			name:     "Erotica",
			rating:   ContentRatingErotica,
			expected: "Explicit",
		},
		{
			name:     "Pornographic",
			rating:   ContentRatingPornographic,
			expected: "Adult",
		},
		{
			name:     "Unknown",
			rating:   "unknown",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.rating.MetronInfoAgeRating())
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
