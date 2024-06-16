package config

import (
	"log/slog"
)

type Config interface {
	GetPort() string
	GetPassWord() string
	GetRootDir() string
	GetRootURl() string
	GetPages() []Page
	HasProvider(Provider) bool
	GetLoggingConfig() LoggingConfig
	GetMaxConcurrentTorrents() int
	GetMaxConcurrentMangadexImages() int
}

type LoggingConfig interface {
	GetLogLevel() slog.Level
	GetSource() bool
	GetHandler() string
	LogHttp() bool
}

type Page interface {
	GetTitle() string
	GetSearchConfig() SearchConfig
}

type SearchConfig interface {
	GetProvider() []Provider
	GetSearchModifiers() map[string]Modifier
	GetRootDirs() []string
	GetCustomRootDir() string
}

type Provider string

const (
	SUKEBEI    Provider = "sukebei"
	NYAA       Provider = "nyaa"
	YTS        Provider = "yts"
	LIME       Provider = "limetorrents"
	SUBSPLEASE Provider = "subsplease"
	MANGADEX   Provider = "mangadex"
)

type Modifier struct {
	Title  string       `yaml:"title" json:"title"`
	Type   ModifierType `yaml:"type" json:"type"`
	Values []Pair       `yaml:"values" json:"values"`
}

type ModifierType string

const (
	DROPDOWN ModifierType = "dropdown"
	MULTI    ModifierType = "multi"
)

func IsValidModifierType(modType ModifierType) bool {
	switch modType {
	case DROPDOWN, MULTI:
		return true
	default:
		return false
	}
}

type Pair struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
}
