package models

import (
	"github.com/lib/pq"
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

	Title         string         `json:"title"`
	Icon          string         `json:"icon"`
	SortValue     int            `json:"sortValue"`
	Providers     pq.Int64Array  `gorm:"type:integer[]" json:"providers"`
	Modifiers     []Modifier     `json:"modifiers"`
	Dirs          pq.StringArray `gorm:"type:text[]" json:"dirs"`
	CustomRootDir string         `json:"customRootDir"`
}

type ModifierType int

const (
	DROPDOWN ModifierType = iota + 1
	MULTI
)

type Modifier struct {
	Model

	PageID int

	Title  string          `json:"title"`
	Type   ModifierType    `json:"type"`
	Key    string          `json:"key"`
	Values []ModifierValue `json:"values"`
	Sort   int
}

type ModifierValue struct {
	Model

	ModifierID int
	Key        string `json:"key"`
	Value      string `json:"value"`
	Default    bool   `json:"default"`
}
