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
	OneHour RefreshFrequency = iota
	HalfDay
	FullDay
	Week
)

func (f RefreshFrequency) AsDuration() time.Duration {
	switch f {
	case OneHour:
		return time.Hour * 1
	case HalfDay:
		return time.Hour * 12
	case FullDay:
		return time.Hour * 24
	case Week:
		return time.Hour * 24 * 7
	}
	panic("invalid refresh frequency")
}
