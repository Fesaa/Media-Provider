package services

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type PreferencesService interface {
	Update(preference models.Preference) error
}

type preferencesService struct {
	pref models.Preferences
	log  zerolog.Logger
}

func PreferenceServiceProvider(pref models.Preferences, log zerolog.Logger) PreferencesService {
	return &preferencesService{
		pref: pref,
		log:  log.With().Str("handler", "preference-service").Logger(),
	}
}

func (p *preferencesService) Update(preference models.Preference) error {
	cur, err := p.pref.GetWithTags()
	if err != nil {
		return err
	}

	newDynastyTags := utils.Filter(preference.DynastyGenreTags, func(tag models.Tag) bool {
		return !models.Tags(cur.DynastyGenreTags).ContainsTag(tag)
	})
	newBlackLists := utils.Filter(preference.BlackListedTags, func(tag models.Tag) bool {
		return !models.Tags(cur.BlackListedTags).ContainsTag(tag)
	})

	preference.BlackListedTags = utils.FlatMapMany(cur.BlackListedTags, newBlackLists)
	preference.DynastyGenreTags = utils.FlatMapMany(cur.DynastyGenreTags, newDynastyTags)

	return p.pref.Update(preference)
}
