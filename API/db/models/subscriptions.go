package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	Model

	Owner            int                     `json:"owner"`
	Provider         Provider                `gorm:"type:int" json:"provider"`
	ContentId        string                  `json:"contentId"`
	RefreshFrequency RefreshFrequency        `gorm:"type:int" json:"refreshFrequency"`
	InfoSql          json.RawMessage         `gorm:"type:jsonb" json:"-"`
	Info             SubscriptionInfo        `gorm:"-" json:"info"`
	Metadata         json.RawMessage         `gorm:"type:jsonb" json:"-"`
	Payload          DownloadRequestMetadata `gorm:"-" json:"metadata"`
}

type DownloadRequestMetadata struct {
	StartImmediately bool                `json:"startImmediately"`
	Extra            map[string][]string `json:"extra,omitempty"`
}

func (s *Subscription) BeforeSave(tx *gorm.DB) (err error) {
	s.Metadata, err = json.Marshal(s.Payload.Extra)
	if err != nil {
		return
	}

	s.InfoSql, err = json.Marshal(s.Info)
	return
}

func (s *Subscription) AfterFind(tx *gorm.DB) (err error) {
	s.Payload.StartImmediately = true

	if s.Metadata != nil {
		err = json.Unmarshal(s.Metadata, &s.Payload.Extra)
		if err != nil {
			return
		}
	}

	if s.InfoSql != nil {
		err = json.Unmarshal(s.InfoSql, &s.Info)
		if err != nil {
			return
		}
	}

	return
}

func (s *Subscription) shouldRefresh(old *Subscription) bool {
	return s.RefreshFrequency != old.RefreshFrequency
}

func (s *Subscription) Normalize(hour int) {
	s.Info.LastCheck = s.normalize(s.Info.LastCheck, hour)
}

func (s *Subscription) normalize(t time.Time, hour int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, time.Local)
}

func (s *Subscription) NextExecution(hour int) time.Time {
	diff := time.Since(s.Info.LastCheck)

	if diff > s.RefreshFrequency.asDuration() {
		next := s.normalize(time.Now(), hour)

		if time.Now().After(next) {
			next = next.Add(time.Hour * 24)
		}

		return next
	}

	next := time.Now().Add(s.RefreshFrequency.asDuration() - diff)
	next = s.normalize(next, hour)

	// Save guard, but should not happen
	if time.Now().After(next) {
		next = next.Add(time.Hour * 24)
	}

	return next
}

type SubscriptionInfo struct {
	Title            string    `json:"title"`
	BaseDir          string    `json:"baseDir"`
	LastCheck        time.Time `json:"lastCheck"`
	LastCheckSuccess bool      `json:"lastCheckSuccess"`
	NextExecution    time.Time `json:"nextExecution"`
}

type RefreshFrequency int

const (
	Day RefreshFrequency = iota + 2
	Week
	Month
)

func (f RefreshFrequency) asDuration() time.Duration {
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
