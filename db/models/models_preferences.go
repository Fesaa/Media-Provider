package models

import (
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

// Add new relations in impl/preferences.go (update & Preferences.GetComplete)

type Preference struct {
	Model

	SubscriptionRefreshHour int                 `json:"subscriptionRefreshHour" validate:"min=0,max=23"`
	LogEmptyDownloads       bool                `json:"logEmptyDownloads" validate:"boolean"`
	CoverFallbackMethod     CoverFallbackMethod `json:"coverFallbackMethod"`
	DynastyGenreTags        []Tag               `json:"dynastyGenreTags" gorm:"many2many:preference_dynasty_genre_tags"`
	BlackListedTags         []Tag               `json:"blackListedTags" gorm:"many2many:preference_black_list_tags"`
	AgeRatingMappings       []AgeRatingMap      `json:"ageRatingMappings" gorm:"many2many:preference_age_rating_mappings"`
}

type CoverFallbackMethod int

const (
	CoverFallbackFirst CoverFallbackMethod = iota
	CoverFallbackLast
	CoverFallbackNone
)

type Tags []Tag

func (tags Tags) ContainsTag(tag Tag) bool {
	return tags.Contains(tag.Name) // Don't trust normalized name
}

func (tags Tags) Contains(tag string) bool {
	nt := utils.Normalize(tag)
	for _, t := range tags {
		if t.Is(tag, nt) {
			return true
		}
	}
	return false
}

type Tag struct {
	Model

	PreferenceID   uint
	AgeRatingMapID uint

	Name           string `json:"name"`
	NormalizedName string `json:"normalizedName"`
}

func (tag *Tag) IsNotNormalized(t string) bool {
	nt := utils.Normalize(t)
	return tag.NormalizedName == nt || tag.Name == t
}

func (tag *Tag) Is(t string, nt string) bool {
	return tag.NormalizedName == nt || tag.Name == t
}

func (tag *Tag) BeforeSave(tx *gorm.DB) error {
	tag.NormalizedName = utils.Normalize(tag.Name)
	return nil
}

type AgeRatingMappings []AgeRatingMap

func (arm AgeRatingMappings) GetAgeRating(tag string) (comicinfo.AgeRating, bool) {
	ageRating := -1
	for _, ageRatingMapping := range arm {
		if !ageRatingMapping.Tag.IsNotNormalized(tag) {
			continue
		}

		ageRating = max(ageRating, comicinfo.AgeRatingIndex[ageRatingMapping.ComicInfoAgeRating])
	}

	if ageRating > -1 {
		return comicinfo.IndexToAgeRating[ageRating], true
	}

	return "", false
}

type AgeRatingMap struct {
	Model

	PreferenceID       uint
	Tag                Tag                 `json:"tag"`
	ComicInfoAgeRating comicinfo.AgeRating `json:"comicInfoAgeRating"`
	// MetronAgeRating    metroninfo.AgeRating `json:"metronAgeRating"`
}
