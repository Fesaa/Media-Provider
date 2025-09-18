package mock

import (
	"github.com/Fesaa/Media-Provider/db/models"
)

type Preferences struct {
	Model models.UserPreferences
}

func (p *Preferences) Get() (*models.UserPreferences, error) {
	return &p.Model, nil
}

func (p *Preferences) GetComplete() (*models.UserPreferences, error) {
	return &p.Model, nil
}

func (p *Preferences) Update(pref models.UserPreferences) error {
	p.Model = pref
	return nil
}

func (p *Preferences) Flush() error {
	return nil
}
