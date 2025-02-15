package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Preference struct {
	gorm.Model

	SubscriptionRefreshHour int            `json:"subscriptionRefreshHour" validate:"min=0,max=23"`
	LogEmptyDownloads       bool           `json:"logEmptyDownloads" validate:"boolean"`
	DynastyGenreTags        pq.StringArray `gorm:"type:string[]" json:"dynastyGenreTags"`
}
