package core

import (
	"io"
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
)

func TestDownloadBase_GetGenreAndTags(t *testing.T) {
	type fields struct {
		genres     []string
		blacklist  []string
		whitelist  []string
		mappings   []models.TagMapping
		includeAll bool
	}

	tests := []struct {
		name   string
		fields fields
		tags   []Tag
		wantG  string
		wantT  string
	}{
		{
			name: "Tag is genre and allowed",
			fields: fields{
				genres: []string{"genre1"},
			},
			tags:  []Tag{NewStringTag("genre1")},
			wantG: "genre1",
			wantT: "",
		},
		{
			name: "Tag is blacklist - excluded from both",
			fields: fields{
				blacklist: []string{"badtag"},
				genres:    []string{"badtag"},
			},
			tags:  []Tag{NewStringTag("badtag")},
			wantG: "",
			wantT: "",
		},
		{
			name: "Tag is whitelist and not genre - becomes tag",
			fields: fields{
				whitelist: []string{"tag1"},
			},
			tags:  []Tag{NewStringTag("tag1")},
			wantG: "",
			wantT: "tag1",
		},
		{
			name: "Tag not matched, but includeAll is true",
			fields: fields{
				includeAll: true,
			},
			tags:  []Tag{NewStringTag("other")},
			wantG: "",
			wantT: "other",
		},
		{
			name: "Mix of genre, tag, blacklist and unmatched",
			fields: fields{
				genres:     []string{"genre1"},
				whitelist:  []string{"tag1"},
				blacklist:  []string{"badtag"},
				includeAll: true,
			},
			tags: []Tag{
				NewStringTag("genre1"),
				NewStringTag("tag1"),
				NewStringTag("badtag"),
				NewStringTag("extra"),
			},
			wantG: "genre1",
			wantT: "tag1, extra",
		},
		{
			name: "test map",
			fields: fields{
				genres:    []string{"genre2"},
				blacklist: []string{"badtag"},
				whitelist: []string{"tag1"},
				mappings: []models.TagMapping{
					{
						OriginTag:      "genre1",
						DestinationTag: "genre2",
					},
				},
			},
			tags: []Tag{
				NewStringTag("genre1"),
			},
			wantG: "genre2",
			wantT: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := req()
			if tt.fields.includeAll {
				r.DownloadMetadata.Extra = map[string][]string{}
				r.DownloadMetadata.Extra[IncludeNotMatchedTagsKey] = []string{"true"}
			}

			base := testBase(t, r, io.Discard, ProviderMock{})
			base.Preference = &models.UserPreferences{
				GenreList:   tt.fields.genres,
				BlackList:   tt.fields.blacklist,
				WhiteList:   tt.fields.whitelist,
				TagMappings: tt.fields.mappings,
			}

			gotG, gotT := base.GetGenreAndTags(t.Context(), tt.tags)
			if gotG != tt.wantG {
				t.Errorf("Genres got = %q, want %q", gotG, tt.wantG)
			}
			if gotT != tt.wantT {
				t.Errorf("Tags got = %q, want %q", gotT, tt.wantT)
			}
		})
	}
}

func TestManga_GetAgeRating(t *testing.T) {
	type test struct {
		name        string
		arm         []models.AgeRatingMapping
		mappings    []models.TagMapping
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
			arm: []models.AgeRatingMapping{
				{
					Tag:                "MyTag",
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
			},
			seriesTags: []Tag{
				NewStringTag("MyTag"),
			},
			chapterTags: []Tag{},
			want:        true,
			wanted:      comicinfo.AgeRatingTeen,
		},
		{
			name: "Chapter Age Rating",
			arm: []models.AgeRatingMapping{
				{
					Tag:                "MyTag",
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
				{
					Tag:                "MyOtherTag",
					ComicInfoAgeRating: comicinfo.AgeRatingMAPlus15,
				},
			},
			seriesTags: []Tag{
				NewStringTag("MyTag"),
			},
			chapterTags: []Tag{
				NewStringTag("MyOtherTag"),
			},
			want:   true,
			wanted: comicinfo.AgeRatingMAPlus15,
		},
		{
			name: "test with map",
			arm: []models.AgeRatingMapping{
				{
					Tag:                "MyTag",
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
			},
			mappings: []models.TagMapping{
				{
					OriginTag:      "MyOtherTag",
					DestinationTag: "MyTag",
				},
			},
			seriesTags: []Tag{
				NewStringTag("MyOtherTag"),
			},
			want:   true,
			wanted: comicinfo.AgeRatingTeen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := testBase(t, req(), io.Discard, ProviderMock{})
			base.Preference = &models.UserPreferences{
				AgeRatingMappings: tt.arm,
				TagMappings:       tt.mappings,
			}

			got, ok := base.GetAgeRating(utils.FlatMapMany(tt.chapterTags, tt.seriesTags))
			if ok != tt.want {
				t.Errorf("m.getAgeRating() got = %v, want %v", ok, tt.want)
			}

			if got != tt.wanted {
				t.Errorf("m.getAgeRating() got = %v, want %v", ok, tt.want)
			}

		})
	}
}
