package mock

import (
	"github.com/Fesaa/Media-Provider/db/models"
)

type Preferences struct {
	model models.Preference
}

func (p *Preferences) Get() (*models.Preference, error) {
	return &p.model, nil
}

func (p *Preferences) GetWithTags() (*models.Preference, error) {
	return &p.model, nil
}

func (p *Preferences) Update(pref models.Preference) error {
	p.model = pref
	return nil
}
