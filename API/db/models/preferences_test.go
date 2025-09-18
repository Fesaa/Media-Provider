package models

import (
	"testing"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"gorm.io/gorm"
)

func TestTags_ContainsTag(t *testing.T) {
	type test struct {
		Name     string
		Tags     Tags
		Tag      Tag
		Expected bool
	}

	tests := []test{
		{
			Name: "Tag exists",
			Tags: Tags{
				{Name: "Action", NormalizedName: "action"},
				{Name: "Adventure", NormalizedName: "adventure"},
			},
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			Expected: true,
		},
		{
			Name: "Tag does not exist",
			Tags: Tags{
				{Name: "Action", NormalizedName: "action"},
				{Name: "Adventure", NormalizedName: "adventure"},
			},
			Tag:      Tag{Name: "Comedy", NormalizedName: "comedy"},
			Expected: false,
		},
		{
			Name:     "Empty tags",
			Tags:     Tags{},
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			Expected: false,
		},
		{
			Name: "Empty tag to search for",
			Tags: Tags{
				{Name: "Action", NormalizedName: "action"},
			},
			Tag:      Tag{Name: "", NormalizedName: ""},
			Expected: false,
		},
		{
			Name: "Tags with different capitalization",
			Tags: Tags{
				{Name: "action", NormalizedName: "action"},
			},
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			Expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Tags.ContainsTag(tt.Tag)
			if result != tt.Expected {
				t.Errorf("Expected %v, got %v", tt.Expected, result)
			}
		})
	}
}

func TestTags_Contains(t *testing.T) {
	type test struct {
		Name     string
		Tags     Tags
		Tag      string
		Expected bool
	}

	tests := []test{
		{
			Name: "Tag exists",
			Tags: Tags{
				{Name: "Action", NormalizedName: "action"},
				{Name: "Adventure", NormalizedName: "adventure"},
			},
			Tag:      "Action",
			Expected: true,
		},
		{
			Name: "Tag does not exist",
			Tags: Tags{
				{Name: "Action", NormalizedName: "action"},
				{Name: "Adventure", NormalizedName: "adventure"},
			},
			Tag:      "Comedy",
			Expected: false,
		},
		{
			Name:     "Empty tags",
			Tags:     Tags{},
			Tag:      "Action",
			Expected: false,
		},
		{
			Name: "Empty tag to search for",
			Tags: Tags{
				{Name: "Action", NormalizedName: "action"},
			},
			Tag:      "",
			Expected: false,
		},
		{
			Name: "Tags with different capitalization",
			Tags: Tags{
				{Name: "action", NormalizedName: "action"},
			},
			Tag:      "Action",
			Expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Tags.Contains(tt.Tag)
			if result != tt.Expected {
				t.Errorf("Expected %v, got %v", tt.Expected, result)
			}
		})
	}
}

func TestTag_IsNotNormalized(t *testing.T) {
	type test struct {
		Name     string
		Tag      Tag
		Input    string
		Expected bool
	}

	tests := []test{
		{
			Name:     "Normalized match",
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			Input:    "action",
			Expected: true,
		},
		{
			Name:     "Name match",
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			Input:    "Action",
			Expected: true,
		},
		{
			Name:     "No match",
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			Input:    "Comedy",
			Expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Tag.IsNotNormalized(tt.Input)
			if result != tt.Expected {
				t.Errorf("Expected %v, got %v", tt.Expected, result)
			}
		})
	}
}

func TestTag_Is(t *testing.T) {
	type test struct {
		Name     string
		Tag      Tag
		InputT   string
		InputNT  string
		Expected bool
	}

	tests := []test{
		{
			Name:     "Normalized match",
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			InputT:   "something",
			InputNT:  "action",
			Expected: true,
		},
		{
			Name:     "Name match",
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			InputT:   "Action",
			InputNT:  "something",
			Expected: true,
		},
		{
			Name:     "No match",
			Tag:      Tag{Name: "Action", NormalizedName: "action"},
			InputT:   "Comedy",
			InputNT:  "comedy",
			Expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Tag.Is(tt.InputT, tt.InputNT)
			if result != tt.Expected {
				t.Errorf("Expected %v, got %v", tt.Expected, result)
			}
		})
	}
}

func TestTag_BeforeSave(t *testing.T) {
	tag := Tag{Name: "Action Adventure"}
	err := tag.BeforeSave(&gorm.DB{})
	if err != nil {
		t.Errorf("Error should be nil, got %v", err)
	}

	if tag.NormalizedName != "actionadventure" {
		t.Errorf("Normalized name should be 'actionadventure', got %v", tag.NormalizedName)
	}
}

func TestAgeRatingMappings_GetAgeRating(t *testing.T) {
	type test struct {
		Name           string
		Mappings       AgeRatingMappings
		Tag            string
		ExpectedRating comicinfo.AgeRating
		ExpectedFound  bool
	}

	tests := []test{
		{
			Name: "Match found",
			Mappings: AgeRatingMappings{
				{Tag: Tag{Name: "Mature", NormalizedName: "mature"}, ComicInfoAgeRating: comicinfo.AgeRatingAdultsOnlyPlus18},
				{Tag: Tag{Name: "Teen", NormalizedName: "teen"}, ComicInfoAgeRating: comicinfo.AgeRatingTeen},
			},
			Tag:            "Mature",
			ExpectedRating: comicinfo.AgeRatingAdultsOnlyPlus18,
			ExpectedFound:  true,
		},
		{
			Name: "Match not found",
			Mappings: AgeRatingMappings{
				{Tag: Tag{Name: "Mature", NormalizedName: "mature"}, ComicInfoAgeRating: comicinfo.AgeRatingAdultsOnlyPlus18},
			},
			Tag:            "Kid",
			ExpectedRating: "",
			ExpectedFound:  false,
		},
		{
			Name: "Multiple matches, highest rating",
			Mappings: AgeRatingMappings{
				{Tag: Tag{Name: "Teen", NormalizedName: "teen"}, ComicInfoAgeRating: comicinfo.AgeRatingTeen},
				{Tag: Tag{Name: "Mature", NormalizedName: "mature"}, ComicInfoAgeRating: comicinfo.AgeRatingAdultsOnlyPlus18},
			},
			Tag:            "Mature",
			ExpectedRating: comicinfo.AgeRatingAdultsOnlyPlus18,
			ExpectedFound:  true,
		},
		{
			Name: "Different capitalization",
			Mappings: AgeRatingMappings{
				{Tag: Tag{Name: "mature", NormalizedName: "mature"}, ComicInfoAgeRating: comicinfo.AgeRatingM},
			},
			Tag:            "Mature",
			ExpectedRating: comicinfo.AgeRatingM,
			ExpectedFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			rating, found := tt.Mappings.GetAgeRating(tt.Tag)
			if found != tt.ExpectedFound {
				t.Errorf("Expected %v, got %v", tt.ExpectedFound, found)
			}

			if rating != tt.ExpectedRating {
				t.Errorf("Expected %v, got %v", tt.ExpectedRating, rating)
			}
		})
	}
}
