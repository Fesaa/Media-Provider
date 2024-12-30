package config

import (
	"github.com/rs/zerolog"
)

// Validate is for API requests, we manually validate stuff on startup

type Config struct {
	Version int    `json:"version"`
	SyncId  int    `json:"sync_id"`
	RootDir string `json:"root_dir"`
	BaseUrl string `json:"base_url"`
	Secret  string `json:"secret"`

	HasUpdatedDB bool `json:"has_updated_db"`

	Logging    Logging     `json:"logging"`
	Downloader Downloader  `json:"downloader"`
	Cache      CacheConfig `json:"cache"`
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
	MaxConcurrentTorrents       int  `json:"max_torrents" validate:"required,number,min=1,max=10"`
	MaxConcurrentMangadexImages int  `json:"max_mangadex_images" validate:"required,number,min=1,max=5"`
	DisableIpv6                 bool `json:"disable_ipv6"`
}

type Logging struct {
	Level   zerolog.Level `json:"level"`
	Source  bool          `json:"source"`
	Handler LogHandler    `json:"handler" validate:"uppercase"`
}

type LogHandler string

const (
	LogHandlerText LogHandler = "TEXT"
	LogHandlerJSON LogHandler = "JSON"
)
