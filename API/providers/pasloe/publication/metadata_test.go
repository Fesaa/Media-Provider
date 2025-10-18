package publication

import (
	"testing"

	"github.com/Fesaa/Media-Provider/utils"
)

func Test_publication_GetCiStatus(t *testing.T) {
	type fields struct {
		SeriesStatus      Status
		TranslationStatus utils.Settable[Status]
		Volumes           []string
		Chapters          []string
		HighestVolume     utils.Settable[float64]
		HighestChapter    utils.Settable[float64]
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  bool
		want2  bool
	}{
		{
			name: "series not completed returns zeros",
			fields: fields{
				SeriesStatus: StatusOngoing,
			},
			want:  0,
			want1: false,
			want2: false,
		},
		{
			name: "translation not completed returns zeros",
			fields: fields{
				SeriesStatus:      StatusCompleted,
				TranslationStatus: utils.NewSettable(StatusOngoing),
			},
			want:  0,
			want1: false,
			want2: false,
		},
		{
			name: "uses HighestVolume when set and matches",
			fields: fields{
				SeriesStatus:  StatusCompleted,
				Volumes:       []string{"1", "2", "3"},
				HighestVolume: utils.NewSettable[float64](3),
			},
			want:  3,
			want1: true,
			want2: true,
		},
		{
			name: "uses HighestVolume when set but doesn't match chapters",
			fields: fields{
				SeriesStatus:  StatusCompleted,
				Volumes:       []string{"1", "2"},
				HighestVolume: utils.NewSettable[float64](3),
			},
			want:  3,
			want1: true,
			want2: false,
		},
		{
			name: "uses HighestChapter when set and matches",
			fields: fields{
				SeriesStatus:   StatusCompleted,
				Chapters:       []string{"1", "2", "10"},
				HighestChapter: utils.NewSettable[float64](10),
			},
			want:  10,
			want1: true,
			want2: true,
		},
		{
			name: "uses HighestChapter when set but doesn't match",
			fields: fields{
				SeriesStatus:   StatusCompleted,
				Chapters:       []string{"1", "2"},
				HighestChapter: utils.NewSettable[float64](10),
			},
			want:  10,
			want1: true,
			want2: false,
		},
		{
			name: "prefers HighestVolume over HighestChapter",
			fields: fields{
				SeriesStatus:   StatusCompleted,
				Volumes:        []string{"1", "2", "3"},
				Chapters:       []string{"1", "2", "10"},
				HighestVolume:  utils.NewSettable[float64](3),
				HighestChapter: utils.NewSettable[float64](10),
			},
			want:  3,
			want1: true,
			want2: true,
		},
		{
			name: "falls back to calculated highest volume",
			fields: fields{
				SeriesStatus: StatusCompleted,
				Volumes:      []string{"1", "2", "5"},
			},
			want:  5,
			want1: true,
			want2: true,
		},
		{
			name: "falls back to calculated highest chapter",
			fields: fields{
				SeriesStatus: StatusCompleted,
				Chapters:     []string{"1", "2", "15"},
			},
			want:  15,
			want1: true,
			want2: true,
		},
		{
			name: "prefers calculated volume over calculated chapter",
			fields: fields{
				SeriesStatus: StatusCompleted,
				Volumes:      []string{"1", "2", "3"},
				Chapters:     []string{"1", "2", "10"},
			},
			want:  3,
			want1: true,
			want2: true,
		},
		{
			name: "no volume or chapter data returns zeros",
			fields: fields{
				SeriesStatus: StatusCompleted,
			},
			want:  0,
			want1: false,
			want2: false,
		},
		{
			name: "translation status not set is allowed",
			fields: fields{
				SeriesStatus:  StatusCompleted,
				Volumes:       []string{"1", "2"},
				HighestVolume: utils.NewSettable[float64](2),
			},
			want:  2,
			want1: true,
			want2: true,
		},
		{
			name: "translation completed is allowed",
			fields: fields{
				SeriesStatus:      StatusCompleted,
				TranslationStatus: utils.NewSettable(StatusCompleted),
				Chapters:          []string{"1", "2", "3"},
				HighestChapter:    utils.NewSettable[float64](3),
			},
			want:  3,
			want1: true,
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &publication{
				series: &Series{
					Status:            tt.fields.SeriesStatus,
					TranslationStatus: tt.fields.TranslationStatus,
					HighestVolume:     tt.fields.HighestVolume,
					HighestChapter:    tt.fields.HighestChapter,
					Chapters:          makeChapters(tt.fields.Volumes, tt.fields.Chapters),
				},
			}
			got, got1, got2 := p.GetCiStatus()
			if got != tt.want {
				t.Errorf("GetCiStatus() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetCiStatus() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetCiStatus() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func makeChapters(volumes, chapters []string) []Chapter {
	maxLen := len(volumes)
	if len(chapters) > maxLen {
		maxLen = len(chapters)
	}

	result := make([]Chapter, maxLen)
	for i := 0; i < maxLen; i++ {
		if i < len(volumes) {
			result[i].Volume = volumes[i]
		}
		if i < len(chapters) {
			result[i].Chapter = chapters[i]
		}
	}
	return result
}
