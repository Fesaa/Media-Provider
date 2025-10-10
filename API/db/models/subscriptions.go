package models

import (
	"encoding/json"
	"time"

	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type Subscription struct {
	Model

	Owner            int                     `json:"owner"`
	Provider         Provider                `gorm:"type:int" json:"provider"`
	ContentId        string                  `json:"contentId"`
	RefreshFrequency RefreshFrequency        `gorm:"type:int" json:"refreshFrequency"`
	Title            string                  `json:"title"`
	BaseDir          string                  `json:"baseDir"`
	LastDownloadDir  string                  `json:"lastDownloadDir"`
	LastCheck        time.Time               `json:"lastCheck"`
	LastCheckSuccess bool                    `json:"lastCheckSuccess"`
	NextExecution    time.Time               `json:"nextExecution"`
	Metadata         json.RawMessage         `gorm:"type:jsonb" json:"-"`
	Payload          DownloadRequestMetadata `gorm:"-" json:"metadata"`
}

type DownloadRequestMetadata struct {
	StartImmediately bool           `json:"startImmediately"`
	Extra            utils.SmartMap `json:"extra,omitempty"`
}

func (s *Subscription) BeforeSave(tx *gorm.DB) (err error) {
	s.Metadata, err = json.Marshal(s.Payload.Extra)
	return
}

func (s *Subscription) AfterFind(tx *gorm.DB) (err error) {
	s.Payload.StartImmediately = true

	if s.Metadata == nil {
		return
	}

	return json.Unmarshal(s.Metadata, &s.Payload.Extra)
}

func (s *Subscription) shouldRefresh(old *Subscription) bool {
	return s.RefreshFrequency != old.RefreshFrequency
}

func (s *Subscription) Normalize(hour int) {
	s.LastCheck = s.normalize(s.LastCheck, hour)
}

func (s *Subscription) normalize(t time.Time, hour int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, time.Local)
}

func (s *Subscription) GetNextExecution(hour int) time.Time {
	diff := time.Since(s.LastCheck)

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
