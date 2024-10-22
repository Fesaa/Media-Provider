package config

import (
	"log/slog"
)

// Validate is for API requests, we manually validate stuff on startup

type Config struct {
	Version int    `json:"version"`
	SyncId  int    `json:"sync_id"`
	RootDir string `json:"root_dir"`
	BaseUrl string `json:"base_url"`
	Secret  string `json:"secret"`

	Logging    Logging     `json:"logging"`
	Downloader Downloader  `json:"downloader"`
	Cache      CacheConfig `json:"cache"`

	Pages []Page `json:"pages"`
}

type CacheConfig struct {
	Type      CacheType `json:"type"`
	RedisAddr string    `json:"redis,omitempty"`
}

type CacheType string

const (
	MEMORY CacheType = "MEMORY"
	REDIS  CacheType = "REDIS"
)

type Downloader struct {
	MaxConcurrentTorrents       int `json:"max_torrents" validate:"required,number,min=1,max=10"`
	MaxConcurrentMangadexImages int `json:"max_mangadex_images" validate:"required,number,min=1,max=5"`
}

type Logging struct {
	Level   slog.Level `json:"level"`
	Source  bool       `json:"source"`
	Handler LogHandler `json:"handler" validate:"uppercase"`
}

type Page struct {
	Title         string              `json:"title" validate:"required,min=3,max=25"`
	Provider      []Provider          `json:"providers" validate:"required,min=1"`
	Modifiers     map[string]Modifier `json:"modifiers"`
	Dirs          []string            `json:"dirs" validate:"required,min=1"`
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
	WEBTOON
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
