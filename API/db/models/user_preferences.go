package models

import (
	"encoding/json"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type CoverFallbackMethod int

const (
	CoverFallbackFirst CoverFallbackMethod = iota
	CoverFallbackLast
	CoverFallbackNone
)

type UserPreferences struct {
	Model

	UserID               int
	LogEmptyDownloads    bool                `json:"logEmptyDownloads" validate:"boolean"`
	ConvertToWebp        bool                `json:"convertToWebp" validate:"boolean"`
	CoverFallbackMethod  CoverFallbackMethod `json:"coverFallbackMethod"`
	GenreList            pq.StringArray      `gorm:"type:text[]" json:"genreList"`
	BlackList            pq.StringArray      `gorm:"type:text[]" json:"blackList"`
	WhiteList            pq.StringArray      `gorm:"type:text[]" json:"whiteList"`
	AgeRatingMappingsSql json.RawMessage     `gorm:"type:jsonb" json:"-"`
	AgeRatingMappings    []AgeRatingMapping  `gorm:"-" json:"ageRatingMappings"`
	TagMappingsSql       json.RawMessage     `gorm:"type:jsonb" json:"-"`
	TagMappings          []TagMapping        `gorm:"-" json:"tagMappings"`
}

func (p *UserPreferences) BeforeSave(tx *gorm.DB) (err error) {
	p.AgeRatingMappingsSql, err = json.Marshal(p.AgeRatingMappings)
	if err != nil {
		return
	}

	p.TagMappingsSql, err = json.Marshal(p.TagMappings)
	return
}

func (p *UserPreferences) AfterFind(tx *gorm.DB) (err error) {
	if p.AgeRatingMappingsSql != nil {
		err = json.Unmarshal(p.AgeRatingMappingsSql, &p.AgeRatingMappings)
		if err != nil {
			return
		}
	}

	if p.TagMappingsSql != nil {
		err = json.Unmarshal(p.TagMappingsSql, &p.TagMappings)
		if err != nil {
			return
		}
	}

	return
}

type AgeRatingMapping struct {
	Tag                string              `json:"tag"`
	ComicInfoAgeRating comicinfo.AgeRating `json:"comicInfoAgeRating"`
	// MetronAgeRating    metroninfo.AgeRating `json:"metronAgeRating"`
}

type TagMapping struct {
	OriginTag      string `json:"originTag"`
	DestinationTag string `json:"destinationTag"`
}
