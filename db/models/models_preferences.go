package models

import "gorm.io/gorm"

type Preference struct {
	gorm.Model

	SubscriptionRefreshHour int  `json:"subscriptionRefreshHour" validate:"min=0,max=23"`
	LogEmptyDownloads       bool `json:"logEmptyDownloads" validate:"boolean"`
}
