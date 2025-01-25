package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

var (
	ErrFailedToLoadPreferences = errors.New("failed to load preferences")
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

func (s *Subscription) Normalize(p Preferences) error {
	pref, err := p.Get()
	if err != nil {
		return ErrFailedToLoadPreferences
	}

	s.Info.LastCheck = s.normalize(s.Info.LastCheck, pref.SubscriptionRefreshHour)

	return nil
}

func (s *Subscription) normalize(t time.Time, hour int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, time.UTC)
}

func (s *Subscription) NextExecution(p Preferences) (time.Time, error) {
	pref, err := p.Get()
	if err != nil {
		return time.Time{}, ErrFailedToLoadPreferences
	}

	diff := time.Since(s.Info.LastCheck)

	if diff > s.RefreshFrequency.AsDuration() {
		next := s.normalize(time.Now(), pref.SubscriptionRefreshHour)

		if time.Now().After(next) {
			next = next.Add(time.Hour * 24)
		}

		return next, nil
	}

	next := time.Now().Add(s.RefreshFrequency.AsDuration() - diff)
	next = s.normalize(next, pref.SubscriptionRefreshHour)

	// Save guard, but should not happen
	if time.Now().After(next) {
		next = next.Add(time.Hour * 24)
	}

	return next, nil
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
