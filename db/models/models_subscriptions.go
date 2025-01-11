package models

import (
	"gorm.io/gorm"
	"time"
)

type Subscription struct {
	gorm.Model

	Provider         Provider         `json:"provider" gorm:"type:int"`
	ContentId        string           `json:"contentId"`
	RefreshFrequency RefreshFrequency `json:"refreshFrequency" gorm:"type:int"`
	Info             SubscriptionInfo `json:"info"`
}

func (s *Subscription) ShouldRefresh(old *Subscription) bool {
	return s.Provider != old.Provider ||
		s.RefreshFrequency != old.RefreshFrequency ||
		s.ContentId != old.ContentId ||
		s.Info.BaseDir != old.Info.BaseDir
}

type SubscriptionInfo struct {
	gorm.Model

	SubscriptionId int

	Title            string    `json:"title"`
	Description      string    `json:"description"`
	BaseDir          string    `json:"baseDir"`
	LastCheck        time.Time `json:"lastCheck"`
	LastCheckSuccess bool      `json:"lastCheckSuccess"`
}

type RefreshFrequency int

const (
	Day RefreshFrequency = iota + 2
	Week
	Month
)

func (f RefreshFrequency) AsDuration() time.Duration {
	switch f {
	case Day:
		return time.Hour * 24
	case Week:
		return time.Hour * 24 * 7
	case Month:
		return time.Hour * 24 * 30
	}
	panic("invalid refresh frequency")
}
