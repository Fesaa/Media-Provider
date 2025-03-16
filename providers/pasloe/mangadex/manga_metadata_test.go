package mangadex

import (
	"bytes"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"io"
	"strings"
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
			m := tempManga(t, req(), io.Discard, &mockRepository{})
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

func TestManga_CIStatus(t *testing.T) {
	type test struct {
		name        string
		volumes     int
		chapters    int
		LastVolume  string
		LastChapter string
		Status      MangaStatus
		wantOk      bool
		wantInLog   []string
		wantCount   int
		isSub       bool
	}

	tests := []test{
		{
			name:   "Mangadex status not completed",
			Status: StatusCancelled,
			wantOk: false,
		},
		{
			name:        "Only chapters",
			Status:      StatusCompleted,
			chapters:    10,
			LastChapter: "10",
			wantOk:      true,
			wantCount:   10,
		},
		{
			name:        "Volumes, all chapters",
			Status:      StatusCompleted,
			volumes:     2,
			chapters:    10,
			LastVolume:  "2",
			LastChapter: "10",
			isSub:       true,
			wantOk:      true,
			wantCount:   2,
			wantInLog:   []string{"Subscription was completed, consider cancelling it"},
		},
		{
			name:        "Volume, missing chapters",
			Status:      StatusCompleted,
			volumes:     2,
			chapters:    8,
			LastVolume:  "2",
			LastChapter: "10",
			isSub:       true,
			wantOk:      true,
			wantCount:   2,
			wantInLog:   []string{"Series ended, but not all chapters could be downloaded or last volume isn't present. English ones missing?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			m := tempManga(t, req(), &buffer, &mockRepository{})
			m.info = &MangaSearchData{
				Attributes: MangaAttributes{
					Status:      tt.Status,
					LastChapter: tt.LastChapter,
					LastVolume:  tt.LastVolume,
				},
			}
			m.Req.IsSubscription = tt.isSub
			m.lastFoundChapter = tt.chapters
			m.lastFoundVolume = tt.volumes

			got, ok := m.getCiStatus()
			if ok != tt.wantOk {
				t.Log(buffer.String())
				t.Errorf("getCiStatus() ok = %v, want %v", ok, tt.wantOk)
			}

			if got != tt.wantCount {
				t.Log(buffer.String())
				t.Errorf("getCiStatus() got = %v, want %v", got, tt.wantCount)
			}

			log := buffer.String()
			for _, logLine := range tt.wantInLog {
				if !strings.Contains(log, logLine) {
					t.Errorf("getCiStatus() got = %v, want %v", log, logLine)
				}
			}
		})
	}
}
