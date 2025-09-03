package services

import (
	"reflect"
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
)

func Test_mergeTags(t *testing.T) {
	type args struct {
		currentTags []models.Tag
		newTags     []models.Tag
	}
	tests := []struct {
		name string
		args args
		want []models.Tag
	}{
		{
			name: "No current",
			args: args{
				currentTags: []models.Tag{},
				newTags: []models.Tag{
					{
						Name:           "tag1",
						NormalizedName: utils.Normalize("tag1"),
					},
					{
						Name:           "tag2",
						NormalizedName: utils.Normalize("tag2"),
					},
				},
			},
			want: []models.Tag{
				{
					Name:           "tag1",
					NormalizedName: utils.Normalize("tag1"),
				},
				{
					Name:           "tag2",
					NormalizedName: utils.Normalize("tag2"),
				},
			},
		},
		{
			name: "No overlapping",
			args: args{
				currentTags: []models.Tag{
					{
						Name:           "tag1",
						NormalizedName: utils.Normalize("tag1"),
					},
				},
				newTags: []models.Tag{
					{
						Name:           "tag2",
						NormalizedName: utils.Normalize("tag2"),
					},
				},
			},
			want: []models.Tag{
				{
					Name:           "tag2",
					NormalizedName: utils.Normalize("tag2"),
				},
			},
		},
		{
			name: "Overlapping",
			args: args{
				currentTags: []models.Tag{
					{
						Name:           "tag1",
						NormalizedName: utils.Normalize("tag1"),
					},
					{
						Name:           "tag2",
						NormalizedName: utils.Normalize("tag2"),
					},
				},
				newTags: []models.Tag{
					{
						Name:           "tAg2",
						NormalizedName: utils.Normalize("tag2"),
					},
				},
			},
			want: []models.Tag{
				{
					Name:           "tag2",
					NormalizedName: utils.Normalize("tag2"),
				},
			},
		},
		{
			name: "No new",
			args: args{
				currentTags: []models.Tag{
					{
						Name:           "tag1",
						NormalizedName: utils.Normalize("tag1"),
					},
					{
						Name:           "tag2",
						NormalizedName: utils.Normalize("tag2"),
					},
				},
				newTags: []models.Tag{
					{
						Name:           "tAg1",
						NormalizedName: utils.Normalize("tag1"),
					},
					{
						Name:           "tag2",
						NormalizedName: utils.Normalize("tag2"),
					},
				},
			},
			want: []models.Tag{
				{
					Name:           "tag1",
					NormalizedName: utils.Normalize("tag1"),
				},
				{
					Name:           "tag2",
					NormalizedName: utils.Normalize("tag2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTags := utils.Map(tt.args.newTags, func(t models.Tag) string {
				return t.Name
			})
			if got := mergeTags(tt.args.currentTags, newTags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
