package models

import "gorm.io/gorm"

type Preference struct {
	gorm.Model

	SubscriptionRefreshHour int `json:"subscriptionRefreshHour" validate:"required,min=0,max=23"`
}
