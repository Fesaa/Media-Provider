package services

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type PreferencesService interface {
	Update(preference payload.PreferencesDto) error
	GetDto() (payload.PreferencesDto, error)
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

func (p *preferencesService) GetDto() (payload.PreferencesDto, error) {
	pref, err := p.pref.GetComplete()
	if err != nil {
		return payload.PreferencesDto{}, err
	}

	toName := func(t models.Tag) string {
		return t.Name
	}

	return payload.PreferencesDto{
		Id:                      pref.ID,
		SubscriptionRefreshHour: pref.SubscriptionRefreshHour,
		LogEmptyDownloads:       pref.LogEmptyDownloads,
		ConvertToWebp:           pref.ConvertToWebp,
		CoverFallbackMethod:     pref.CoverFallbackMethod,
		DynastyGenreTags:        utils.Map(pref.DynastyGenreTags, toName),
		BlackListedTags:         utils.Map(pref.BlackListedTags, toName),
		WhiteListedTags:         utils.Map(pref.WhiteListedTags, toName),
		AgeRatingMappings: utils.Map(pref.AgeRatingMappings, func(t models.AgeRatingMap) payload.AgeRatingMapDto {
			return payload.AgeRatingMapDto{
				Id:                 t.ID,
				Tag:                toName(t.Tag),
				ComicInfoAgeRating: t.ComicInfoAgeRating,
			}
		}),
		TagMappings: utils.Map(pref.TagMappings, func(t models.TagMap) payload.TagMapDto {
			return payload.TagMapDto{
				Id:     t.ID,
				Origin: toName(t.Origin),
				Dest:   toName(t.Dest),
			}
		}),
	}, nil
}

func (p *preferencesService) Update(preference payload.PreferencesDto) error {
	cur, err := p.pref.GetComplete()
	if err != nil {
		return err
	}

	if cur.SubscriptionRefreshHour != preference.SubscriptionRefreshHour {
		if err = p.subscriptionService.UpdateTask(preference.SubscriptionRefreshHour); err != nil {
			return err
		}
	}

	cur.SubscriptionRefreshHour = preference.SubscriptionRefreshHour
	cur.LogEmptyDownloads = preference.LogEmptyDownloads
	cur.ConvertToWebp = preference.ConvertToWebp
	cur.CoverFallbackMethod = preference.CoverFallbackMethod

	cur.BlackListedTags = mergeTags(cur.BlackListedTags, preference.BlackListedTags)
	cur.DynastyGenreTags = mergeTags(cur.DynastyGenreTags, preference.DynastyGenreTags)
	cur.WhiteListedTags = mergeTags(cur.WhiteListedTags, preference.WhiteListedTags)
	cur.AgeRatingMappings = utils.Map(preference.AgeRatingMappings, func(agm payload.AgeRatingMapDto) models.AgeRatingMap {
		return models.AgeRatingMap{
			Tag: models.Tag{
				Name:           agm.Tag,
				NormalizedName: utils.Normalize(agm.Tag),
			},
			ComicInfoAgeRating: agm.ComicInfoAgeRating,
		}
	})
	cur.TagMappings = utils.Map(preference.TagMappings, func(tm payload.TagMapDto) models.TagMap {
		return models.TagMap{
			Origin: models.Tag{
				Name:           tm.Origin,
				NormalizedName: utils.Normalize(tm.Origin),
			},
			Dest: models.Tag{
				Name:           tm.Dest,
				NormalizedName: utils.Normalize(tm.Dest),
			},
		}
	})

	if err = p.pref.Update(*cur); err != nil {
		return err
	}

	// Reset preference cache
	return p.pref.Flush()
}

func mergeTags(currentTags []models.Tag, newTags []string) []models.Tag {
	normalizedNames := make(map[string]string)
	newTagMap := make(map[string]struct{})
	for _, tag := range newTags {
		nt := utils.Normalize(tag)
		newTagMap[nt] = struct{}{}
		normalizedNames[tag] = nt
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
		_, alreadyExists := added[normalizedNames[newTag]]
		if !alreadyExists {
			mergedTags = append(mergedTags, models.Tag{
				Name:           newTag,
				NormalizedName: normalizedNames[newTag],
			})
			added[normalizedNames[newTag]] = true
		}
	}

	return mergedTags
}
