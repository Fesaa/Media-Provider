package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"strings"
)

func InsertDefaultPreferences(db *gorm.DB, log zerolog.Logger) error {
	pref := models.Preference{
		SubscriptionRefreshHour: 0,
	}
	return db.Save(&pref).Error
}

func MigrateTags(db *gorm.DB, log zerolog.Logger) error {
	if getCurrentVersion(db) != "" {
		log.Trace().Msg("Skipping changes, Media-Provider installed after changes are needed")
		return nil
	}
	var blackList pq.StringArray
	var dynasty pq.StringArray

	err := db.Raw("SELECT dynasty_genre_tags, black_listed_tags FROM preferences").Row().Scan(&dynasty, &blackList)
	if err != nil {
		if strings.Contains(err.Error(), "no such column") {
			return nil
		}

		return err
	}

	var p models.Preference
	if err = db.First(&p).Error; err != nil {
		return err
	}

	toTag := func(s string) models.Tag {
		return models.Tag{
			PreferenceID:   p.ID,
			Name:           s,
			NormalizedName: utils.Normalize(s),
		}
	}

	p.DynastyGenreTags = utils.Map(dynasty, toTag)
	p.BlackListedTags = utils.Map(blackList, toTag)
	if err = db.Save(&p).Error; err != nil {
		return err
	}

	err = db.Exec("ALTER TABLE preferences DROP COLUMN black_listed_tags").Error
	if err != nil {
		return err
	}

	err = db.Exec("ALTER TABLE preferences DROP COLUMN dynasty_genre_tags").Error
	if err != nil {
		return err
	}

	return nil
}
