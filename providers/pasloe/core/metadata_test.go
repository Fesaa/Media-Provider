package core

import (
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"testing"
)

type preferences struct {
	m *models.Preference
}

func (p preferences) Get() (*models.Preference, error) {
	return p.m, nil
}

func (p preferences) GetComplete() (*models.Preference, error) {
	return p.m, nil
}

func (p preferences) Update(pref models.Preference) error {
	return nil
}

func (p preferences) Flush() error {
	return nil
}

func TestDownloadBase_GetGenreAndTags(t *testing.T) {
	type fields struct {
		genres     models.Tags
		blacklist  models.Tags
		whitelist  models.Tags
		mappings   []models.TagMap
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
				genres: models.Tags{models.NewTag("genre1")},
			},
			tags:  []Tag{NewStringTag("genre1")},
			wantG: "genre1",
			wantT: "",
		},
		{
			name: "Tag is blacklist - excluded from both",
			fields: fields{
				blacklist: models.Tags{models.NewTag("badtag")},
				genres:    models.Tags{models.NewTag("badtag")},
			},
			tags:  []Tag{NewStringTag("badtag")},
			wantG: "",
			wantT: "",
		},
		{
			name: "Tag is whitelist and not genre - becomes tag",
			fields: fields{
				whitelist: models.Tags{models.NewTag("tag1")},
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
				genres:     models.Tags{models.NewTag("genre1")},
				whitelist:  models.Tags{models.NewTag("tag1")},
				blacklist:  models.Tags{models.NewTag("badtag")},
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
				genres:    models.Tags{models.NewTag("genre2")},
				blacklist: models.Tags{models.NewTag("badtag")},
				whitelist: models.Tags{models.NewTag("tag1")},
				mappings: []models.TagMap{
					{
						Origin: models.NewTag("genre1"),
						Dest:   models.NewTag("genre2"),
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
			base.preferences = preferences{
				m: &models.Preference{
					DynastyGenreTags: tt.fields.genres,
					BlackListedTags:  tt.fields.blacklist,
					WhiteListedTags:  tt.fields.whitelist,
					TagMappings:      tt.fields.mappings,
				},
			}

			gotG, gotT := base.GetGenreAndTags(tt.tags)
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
		arm         []models.AgeRatingMap
		mappings    []models.TagMap
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
				NewStringTag("MyTag"),
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
			arm: []models.AgeRatingMap{
				{
					Tag: models.Tag{
						Name:           "MyTag",
						NormalizedName: "mytag",
					},
					ComicInfoAgeRating: comicinfo.AgeRatingTeen,
				},
			},
			mappings: []models.TagMap{
				{
					Origin: models.NewTag("MyOtherTag"),
					Dest:   models.NewTag("MyTag"),
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
			base.Preference = &models.Preference{
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
