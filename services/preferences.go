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
	subscriptionService SubscriptionService
	pref                models.Preferences
	log                 zerolog.Logger
}

func PreferenceServiceProvider(subscriptionService SubscriptionService, pref models.Preferences, log zerolog.Logger) PreferencesService {
	return &preferencesService{
		pref:                pref,
		subscriptionService: subscriptionService,
		log:                 log.With().Str("handler", "preference-service").Logger(),
	}
}

func (p *preferencesService) Update(preference models.Preference) error {
	cur, err := p.pref.GetComplete()
	if err != nil {
		return err
	}

	if cur.SubscriptionRefreshHour != preference.SubscriptionRefreshHour {
		if err = p.subscriptionService.UpdateTask(preference.SubscriptionRefreshHour); err != nil {
			return err
		}
	}

	// TODO: Could we pass all tags as current? So they share tags?
	preference.BlackListedTags = mergeTags(cur.BlackListedTags, preference.BlackListedTags)
	preference.DynastyGenreTags = mergeTags(cur.DynastyGenreTags, preference.DynastyGenreTags)
	preference.WhiteListedTags = mergeTags(cur.WhiteListedTags, preference.WhiteListedTags)

	if err = p.pref.Update(preference); err != nil {
		return err
	}

	// Reset preference cache
	return p.pref.Flush()
}

func mergeTags(currentTags, newTags []models.Tag) []models.Tag {
	normalizedNames := make(map[string]string)
	newTagMap := make(map[string]struct{})
	for _, tag := range newTags {
		nt := utils.Normalize(tag.Name)
		newTagMap[nt] = struct{}{}
		normalizedNames[tag.Name] = nt
	}

	mergedTags := make([]models.Tag, 0)

	// Remove deleted tags
	added := map[string]bool{}
	for _, currentTag := range currentTags {
		if _, exists := newTagMap[currentTag.NormalizedName]; exists {
			mergedTags = append(mergedTags, currentTag)
			added[currentTag.NormalizedName] = true
		}
	}

	// Add new tags
	for _, newTag := range newTags {
		_, alreadyExists := added[normalizedNames[newTag.Name]]
		if !alreadyExists {
			mergedTags = append(mergedTags, newTag)
			added[normalizedNames[newTag.Name]] = true
		}
	}

	return mergedTags
}
