package models

import (
	"encoding/json"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Provider int

const (
	NYAA Provider = iota + 2
	YTS
	LIME
	SUBSPLEASE
	MANGADEX
	WEBTOON
	DYNASTY
	BATO

	MinProvider = NYAA
	MaxProvider = BATO
)

func (p Provider) String() string {
	switch p {
	case NYAA:
		return "Nyaa"
	case YTS:
		return "YTS"
	case LIME:
		return "Lime"
	case SUBSPLEASE:
		return "SubsPlease"
	case MANGADEX:
		return "MangaDex"
	case WEBTOON:
		return "Webtoon"
	case DYNASTY:
		return "Dynasty"
	case BATO:
		return "Bato"
	default:
		return "Unknown Provider"
	}
}

type Page struct {
	Model

	Title         string          `json:"title"`
	Icon          string          `json:"icon"`
	SortValue     int             `json:"sortValue"`
	Providers     pq.Int64Array   `gorm:"type:integer[]" json:"providers"`
	ModifierData  json.RawMessage `gorm:"type:jsonb" json:"-"`
	Modifiers     []Modifier      `gorm:"-" json:"modifiers"`
	Dirs          pq.StringArray  `gorm:"type:string[]" json:"dirs"`
	CustomRootDir string          `json:"customRootDir"`
}

func (p *Page) BeforeSave(tx *gorm.DB) (err error) {
	p.ModifierData, err = json.Marshal(p.Modifiers)
	return
}

func (p *Page) AfterFind(tx *gorm.DB) (err error) {
	if p.ModifierData == nil {
		p.Modifiers = []Modifier{}
		return
	}

	return json.Unmarshal(p.ModifierData, &p.Modifiers)
}

type ModifierType int

const (
	DROPDOWN ModifierType = iota + 1
	MULTI
	Switch
)

type Modifier struct {
	Title  string          `json:"title"`
	Type   ModifierType    `json:"type"`
	Key    string          `json:"key"`
	Values []ModifierValue `json:"values"`
	Sort   int             `json:"sort"`
}

type ModifierValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Default bool   `json:"default"`
}
