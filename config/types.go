package config

import "log/slog"

type Config interface {
	GetPort() string
	GetPassWord() string
	GetRootDir() string
	GetRootURl() string
	GetPages() Pages
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
	GetProvider() SearchProvider
	GetCategories() []Category
	GetSortBys() []SortBy
	GetRootDirs() []string
	GetCustomRootDir() string
}

type SearchProvider string

const (
	NYAA       SearchProvider = "nyaa"
	YTS        SearchProvider = "yts"
	LIME       SearchProvider = "limetorrents"
	SUBSPLEASE SearchProvider = "subsplease"
)

type Category Pair
type SortBy Pair

type Pair struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
}
