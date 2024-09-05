package config

import "log/slog"

// Validate is for API requests, we manually validate stuff on startup

type Config struct {
	SyncId   int    `json:"sync_id"`
	Port     string `json:"port" validate:"required,number"`
	Password string `json:"password" validate:"required"`
	RootDir  string `json:"root_dir"`
	BaseUrl  string `json:"base_url"`

	Logging    Logging    `json:"logging"`
	Downloader Downloader `json:"downloader"`

	Pages []Page `json:"pages"`
}

type Downloader struct {
	MaxConcurrentTorrents       int `json:"max_torrents" validate:"required,number,min=1,max=10"`
	MaxConcurrentMangadexImages int `json:"max_mangadex_images" validate:"required,number,min=1,max=5"`
}

type Logging struct {
	Level   slog.Level `json:"level"`
	Source  bool       `json:"source"`
	Handler LogHandler `json:"handler" validate:"uppercase"`
	LogHttp bool       `json:"log_http"`
}

type Page struct {
	Title         string              `json:"title"`
	Provider      []Provider          `json:"provider"`
	Modifiers     map[string]Modifier `json:"modifiers"`
	Dirs          []string            `json:"dirs"`
	CustomRootDir string              `json:"custom_root_dir"`
}

type Provider int

const (
	SUKEBEI Provider = iota + 1
	NYAA
	YTS
	LIME
	SUBSPLEASE
	MANGADEX
)

type ModifierType int

const (
	DROPDOWN ModifierType = iota + 1
	MULTI
)

func IsValidModifierType(modType ModifierType) bool {
	switch modType {
	case DROPDOWN, MULTI:
		return true
	default:
		return false
	}
}

type Modifier struct {
	Title  string            `yaml:"title" json:"title"`
	Type   ModifierType      `yaml:"type" json:"type"`
	Values map[string]string `yaml:"values" json:"values"`
}

type LogHandler string

const (
	LogHandlerText LogHandler = "TEXT"
	LogHandlerJSON LogHandler = "JSON"
)
