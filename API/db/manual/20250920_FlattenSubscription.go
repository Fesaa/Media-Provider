package manual

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type oldSubscription struct {
	models.Model

	InfoSql json.RawMessage `gorm:"type:jsonb" json:"-"`
	Info    infoSql         `gorm:"-" json:"info"`
}

func (s *oldSubscription) AfterFind(tx *gorm.DB) (err error) {
	if s.InfoSql != nil {
		err = json.Unmarshal(s.InfoSql, &s.Info)
		if err != nil {
			return
		}
	}

	return
}

type infoSql struct {
	Title            string    `json:"title"`
	BaseDir          string    `json:"baseDir"`
	LastCheck        time.Time `json:"lastCheck"`
	LastCheckSuccess bool      `json:"lastCheckSuccess"`
	NextExecution    time.Time `json:"nextExecution"`
}

func FlattenSubscription(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	if !db.WithContext(ctx).Migrator().HasColumn("subscriptions", "info_sql") {
		return nil
	}

	var oldSubs []oldSubscription
	var newSubs []models.Subscription

	if err := db.WithContext(ctx).Table("subscriptions").Find(&oldSubs).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Find(&newSubs).Error; err != nil {
		return err
	}

	newSubsDict := make(map[int]*models.Subscription)
	for _, sub := range newSubs {
		newSubsDict[sub.ID] = &sub
	}

	for _, oldSub := range oldSubs {
		newSub, ok := newSubsDict[oldSub.ID]
		if !ok {
			log.Warn().Int("id", oldSub.ID).Msg("subscription not found in database")
			continue
		}

		newSub.Title = oldSub.Info.Title
		newSub.BaseDir = oldSub.Info.BaseDir
		newSub.LastCheck = oldSub.Info.LastCheck
		newSub.LastCheckSuccess = oldSub.Info.LastCheckSuccess
		newSub.NextExecution = oldSub.Info.NextExecution
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sub := range newSubsDict {
			if err := tx.Save(sub).Error; err != nil {
				return err
			}
		}

		// Sqlite has a NPE
		if config.DbProvider == "postgres" {
			return tx.Migrator().DropColumn("subscriptions", "info_sql")
		}

		return nil
	})
}
