package config

import (
	"log/slog"
)

type Config interface {
	GetPort() string
	GetPassWord() string
	GetRootDir() string
	GetRootURl() string
	GetPages() Pages
	HasProvider(Provider) bool
	GetLoggingConfig() LoggingConfig
}

type LoggingConfig interface {
	GetLogLevel() slog.Level
	GetSource() bool
	GetHandler() string
}

type Pages []Page

type Page interface {
	GetTitle() string
	GetSearchConfig() SearchConfig
}

type SearchConfig interface {
	GetProvider() Provider
	GetSearchModifiers() map[string]Modifier
	GetRootDirs() []string
	GetCustomRootDir() string
}

type Provider string

const (
	NYAA       Provider = "nyaa"
	YTS        Provider = "yts"
	LIME       Provider = "limetorrents"
	SUBSPLEASE Provider = "subsplease"
	MANGADEX   Provider = "mangadex"
)

type Modifier struct {
	Title  string `yaml:"title" json:"title"`
	Multi  bool   `yaml:"multi" json:"multi"`
	Values []Pair `yaml:"values" json:"values"`
}

type Pair struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
}
