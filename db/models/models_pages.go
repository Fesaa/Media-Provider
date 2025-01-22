package models

import (
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
	default:
		return "Unknown Provider"
	}
}

type Page struct {
	gorm.Model

	Title         string         `json:"title"`
	Icon          string         `json:"icon"`
	SortValue     int            `json:"sortValue"`
	Providers     pq.Int64Array  `gorm:"type:integer[]" json:"providers"`
	Modifiers     []Modifier     `json:"modifiers"`
	Dirs          pq.StringArray `gorm:"type:string[]" json:"dirs"`
	CustomRootDir string         `json:"custom_root_dir"`
}

type ModifierType int

const (
	DROPDOWN ModifierType = iota + 1
	MULTI
)

type Modifier struct {
	gorm.Model

	PageID uint

	Title  string          `json:"title"`
	Type   ModifierType    `json:"type"`
	Key    string          `json:"key"`
	Values []ModifierValue `json:"values"`
}

type ModifierValue struct {
	gorm.Model

	ModifierID uint
	Key        string `json:"key"`
	Value      string `json:"value"`
}
