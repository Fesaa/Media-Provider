package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Page struct {
	gorm.Model

	Title         string         `json:"title"`
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

	PageId uint

	Title  string          `json:"title"`
	Type   ModifierType    `json:"type"`
	Key    string          `json:"key"`
	Values []ModifierValue `json:"values"`
}

type ModifierValue struct {
	gorm.Model

	ModifierId uint
	Key        string `json:"key"`
	Value      string `json:"value"`
}
