package models

import (
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type Preference struct {
	gorm.Model

	SubscriptionRefreshHour int                 `json:"subscriptionRefreshHour" validate:"min=0,max=23"`
	LogEmptyDownloads       bool                `json:"logEmptyDownloads" validate:"boolean"`
	CoverFallbackMethod     CoverFallbackMethod `json:"coverFallbackMethod"`
	DynastyGenreTags        []Tag               `json:"dynastyGenreTags" gorm:"many2many:preference_dynasty_genre_tags"`
	BlackListedTags         []Tag               `json:"blackListedTags" gorm:"many2many:preference_black_list_tags"`
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
		if t.NormalizedName == nt || t.Name == tag {
			return true
		}
	}
	return false
}

type Tag struct {
	gorm.Model

	PreferenceID uint

	Name           string `json:"name"`
	NormalizedName string `json:"normalizedName"`
}

func (tag *Tag) BeforeSave(tx *gorm.DB) error {
	tag.NormalizedName = utils.Normalize(tag.Name)
	return nil
}
