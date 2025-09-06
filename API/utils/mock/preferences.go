package mock

import (
	"github.com/Fesaa/Media-Provider/db/models"
)

type Preferences struct {
	Model models.Preference
}

func (p *Preferences) Get() (*models.Preference, error) {
	return &p.Model, nil
}

func (p *Preferences) GetComplete() (*models.Preference, error) {
	return &p.Model, nil
}

func (p *Preferences) Update(pref models.Preference) error {
	p.Model = pref
	return nil
}

func (p *Preferences) Flush() error {
	return nil
}
