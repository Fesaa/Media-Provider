package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
)

type PreferencesDto struct {
	Id                      int                        `json:"id"`
	SubscriptionRefreshHour int                        `json:"subscriptionRefreshHour" validate:"min=0,max=23"`
	LogEmptyDownloads       bool                       `json:"logEmptyDownloads" validate:"boolean"`
	ConvertToWebp           bool                       `json:"convertToWebp" validate:"boolean"`
	CoverFallbackMethod     models.CoverFallbackMethod `json:"coverFallbackMethod"`
	DynastyGenreTags        []string                   `json:"dynastyGenreTags" gorm:"many2many:preference_dynasty_genre_tags"`
	BlackListedTags         []string                   `json:"blackListedTags" gorm:"many2many:preference_black_list_tags"`
	WhiteListedTags         []string                   `json:"whiteListedTags" gorm:"many2many:preference_white_list_tags"`
	AgeRatingMappings       []AgeRatingMapDto          `json:"ageRatingMappings" gorm:"many2many:preference_age_rating_mappings"`
	TagMappings             []TagMapDto                `json:"tagMappings"`
}

type AgeRatingMapDto struct {
	Id                 int                 `json:"id"`
	Tag                string              `json:"tag"`
	ComicInfoAgeRating comicinfo.AgeRating `json:"comicInfoAgeRating"`
}

type TagMapDto struct {
	Id     int    `json:"id"`
	Origin string `json:"origin"`
	Dest   string `json:"dest"`
}
