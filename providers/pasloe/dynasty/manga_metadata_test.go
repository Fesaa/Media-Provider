package dynasty

import (
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"io"
	"testing"
)

func TestManga_GetAgeRating(t *testing.T) {
	type test struct {
		name        string
		arm         []models.AgeRatingMap
		seriesTags  []Tag
		chapterTags []Tag
		want        bool
		wanted      comicinfo.AgeRating
	}

	tests := []test{
		{
			name: "No age rating",
			want: false,
		},
		{
			name: "Series Age Rating",
			arm: []models.AgeRatingMap{
				{
					Tag: models.Tag{
						Name:           "MyTag",
						NormalizedName: "mytag",
					},
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
			},
			seriesTags: []Tag{
				{
					DisplayName: "MyTag",
				},
			},
			chapterTags: []Tag{},
			want:        true,
			wanted:      comicinfo.AgeRatingTeen,
		},
		{
			name: "Chapter Age Rating",
			arm: []models.AgeRatingMap{
				{
					Tag: models.Tag{
						Name:           "MyTag",
						NormalizedName: "mytag",
					},
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
				{
					Tag: models.Tag{
						Name:           "MyOtherTag",
						NormalizedName: "myothertag",
					},
					ComicInfoAgeRating: comicinfo.AgeRatingMAPlus15,
				},
			},
			seriesTags: []Tag{
				{
					DisplayName: "MyTag",
				},
			},
			chapterTags: []Tag{
				{
					DisplayName: "MyOtherTag",
				},
			},
			want:   true,
			wanted: comicinfo.AgeRatingMAPlus15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tempManga(t, req(), io.Discard, &mockRepository{})
			m.Preference = &models.Preference{
				AgeRatingMappings: tt.arm,
			}
			m.seriesInfo = &Series{
				Tags: tt.seriesTags,
			}
			chp := chapter()
			chp.Tags = tt.chapterTags

			got, ok := m.getAgeRating(chp)
			if ok != tt.want {
				t.Errorf("m.getAgeRating() got = %v, want %v", ok, tt.want)
			}

			if got != tt.wanted {
				t.Errorf("m.getAgeRating() got = %v, want %v", ok, tt.want)
			}

		})
	}
}
