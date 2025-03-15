package impl

import (
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"testing"
)

func seedPreferences(t *testing.T, p models.Preferences) {
	t.Helper()
	err := (p.(*preferences)).db.Save(&models.Preference{
		SubscriptionRefreshHour: 5,
		LogEmptyDownloads:       true,
		CoverFallbackMethod:     models.CoverFallbackLast,
		DynastyGenreTags: []models.Tag{
			{Name: "MyTag", NormalizedName: "mytag"},
		},
		BlackListedTags: []models.Tag{
			{Name: "MyOtherTag", NormalizedName: "myothertag"},
		},
		AgeRatingMappings: []models.AgeRatingMap{
			{
				Tag: models.Tag{
					Name:           "MyAgeRatingTag",
					NormalizedName: "myageratingtag",
				},
				ComicInfoAgeRating: comicinfo.AgeRatingMAPlus15,
			},
		},
	}).Error

	if err != nil {
		t.Fatal(err)
	}
}

func TestPreferences_Get(t *testing.T) {
	p := Preferences(databaseHelper(t))
	seedPreferences(t, p)

	got, err := p.Get()
	if err != nil {
		t.Fatal(err)
	}

	if got.SubscriptionRefreshHour != 5 {
		t.Errorf("expected SubscriptionRefreshHour to be 5, got %d", got.SubscriptionRefreshHour)
	}

	if got.CoverFallbackMethod != models.CoverFallbackLast {
		t.Errorf("expected CoverFallbackMethod to be 'last', got %v", got.CoverFallbackMethod)
	}

	if !got.LogEmptyDownloads {
		t.Errorf("expected LogEmptyDownloads to be true, got %t", got.LogEmptyDownloads)
	}

	// Get shouldn't load Associations
	if len(got.BlackListedTags) > 0 {
		t.Errorf("expected BlackListedTags to be empty, got %d", len(got.BlackListedTags))
	}

	if len(got.AgeRatingMappings) > 0 {
		t.Errorf("expected AgeRatingMappings to be empty, got %d", len(got.AgeRatingMappings))
	}

	if len(got.DynastyGenreTags) > 0 {
		t.Errorf("expected BlackListedTags to be empty, got %d", len(got.BlackListedTags))
	}
}

func TestPreferences_GetComplete(t *testing.T) {
	p := Preferences(databaseHelper(t))
	seedPreferences(t, p)

	got, err := p.GetComplete()
	if err != nil {
		t.Fatal(err)
	}

	if len(got.BlackListedTags) == 0 {
		t.Errorf("expected BlackListedTags to not be empty, got %d", len(got.BlackListedTags))
	}

	if got.BlackListedTags[0].Name != "MyOtherTag" {
		t.Errorf("expected BlackListedTag[0].Name to be MyOtherTag, got %s", got.BlackListedTags[0].Name)
	}

	if len(got.DynastyGenreTags) == 0 {
		t.Errorf("expected DynastyGenreTags to not be empty, got %d", len(got.DynastyGenreTags))
	}

	if got.DynastyGenreTags[0].Name != "MyTag" {
		t.Errorf("expected DynastyGenreTags[0].Name to be MyTag, got %s", got.DynastyGenreTags[0].Name)
	}

	if len(got.AgeRatingMappings) == 0 {
		t.Errorf("expected AgeRatingMappings to not be empty, got %d", len(got.AgeRatingMappings))
	}

	if got.AgeRatingMappings[0].Tag.Name != "MyAgeRatingTag" {
		t.Errorf("expected AgeRatingMappings[0].Tag.Name to be MyAgeRatingTag, got %s", got.AgeRatingMappings[0].Tag.Name)
	}

	if got.AgeRatingMappings[0].ComicInfoAgeRating != comicinfo.AgeRatingMAPlus15 {
		t.Errorf("expected AgeRatingMappings[0].ComicInfoAgeRating to be AgeRatingMAPlus15, got %v", got.AgeRatingMappings[0].ComicInfoAgeRating)
	}
}

func TestPreferences_Update(t *testing.T) {
	p := Preferences(databaseHelper(t))
	seedPreferences(t, p)

	pref, err := p.GetComplete()
	if err != nil {
		t.Fatal(err)
	}

	// Update depth 0
	pref.SubscriptionRefreshHour = 2
	if err = p.Update(*pref); err != nil {
		t.Fatal(err)
	}

	pref, err = p.GetComplete()
	if err != nil {
		t.Fatal(err)
	}

	if pref.SubscriptionRefreshHour != 2 {
		t.Errorf("expected SubscriptionRefreshHour to be 2, got %d", pref.SubscriptionRefreshHour)
	}

	// Update depth 0, "empty field"
	pref.LogEmptyDownloads = false
	if err = p.Update(*pref); err != nil {
		t.Fatal(err)
	}

	pref, err = p.GetComplete()
	if err != nil {
		t.Fatal(err)
	}

	if pref.LogEmptyDownloads {
		t.Errorf("expected LogEmptyDownloads to be false, got %t", pref.LogEmptyDownloads)
	}

	// Update depth 1
	pref.BlackListedTags = append(pref.BlackListedTags, models.Tag{
		Name:           "MyTag2",
		NormalizedName: "mytag2",
	})

	if err = p.Update(*pref); err != nil {
		t.Fatal(err)
	}

	pref, err = p.GetComplete()
	if err != nil {
		t.Fatal(err)
	}

	if len(pref.BlackListedTags) != 2 {
		t.Errorf("expected BlackListedTags to be 2, got %d", len(pref.BlackListedTags))
	}

	// Update depth 1, delete
	pref.AgeRatingMappings = make([]models.AgeRatingMap, 0)
	if err = p.Update(*pref); err != nil {
		t.Fatal(err)
	}

	pref, err = p.GetComplete()
	if err != nil {
		t.Fatal(err)
	}

	if len(pref.AgeRatingMappings) != 0 {
		t.Errorf("expected BlackListedTags to be 0, got %d", len(pref.AgeRatingMappings))
	}
}
