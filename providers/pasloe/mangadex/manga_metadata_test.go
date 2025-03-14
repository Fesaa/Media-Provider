package mangadex

import (
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"io"
	"testing"
)

func TestManga_AgeRating(t *testing.T) {
	type test struct {
		name     string
		arm      []models.AgeRatingMap
		tags     []TagData
		mangadex ContentRating
		wanted   comicinfo.AgeRating
	}

	tests := []test{
		{
			name:     "Not overwrites",
			arm:      nil,
			mangadex: ContentRatingSafe,
			wanted:   comicinfo.AgeRatingEveryone,
		},
		{
			name: "Mangadex highest",
			arm: []models.AgeRatingMap{
				{
					Tag: models.Tag{
						Name:           "MyTag",
						NormalizedName: "mytag",
					},
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
			},
			tags: []TagData{
				{
					Attributes: TagAttributes{
						Name: map[string]string{
							"en": "MyTag",
						},
					},
				},
			},
			mangadex: ContentRatingPornographic,
			wanted:   comicinfo.AgeRatingAdultsOnlyPlus18,
		},
		{
			name: "Should overwrite",
			arm: []models.AgeRatingMap{
				{
					Tag: models.Tag{
						Name:           "MyTag",
						NormalizedName: "mytag",
					},
					ComicInfoAgeRating: comicinfo.AgeRatingMAPlus15,
				},
			},
			tags: []TagData{
				{
					Attributes: TagAttributes{
						Name: map[string]string{
							"en": "MyTag",
						},
					},
				},
			},
			mangadex: ContentRatingSafe,
			wanted:   comicinfo.AgeRatingMAPlus15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tempManga(t, req(), io.Discard)
			m.Preference = &models.Preference{
				AgeRatingMappings: tt.arm,
			}
			m.info = &MangaSearchData{
				Attributes: MangaAttributes{
					ContentRating: tt.mangadex,
					Tags:          tt.tags,
				},
			}

			got := m.getAgeRating()
			if got != tt.wanted {
				t.Errorf("getAgeRating() got = %v, want %v", got, tt.wanted)
			}
		})
	}

}
