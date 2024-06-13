package config

import (
	"github.com/Fesaa/Media-Provider/providers"
	"log/slog"
)

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
	GetProvider() providers.SearchProvider
	GetCategories() []Category
	GetSortBys() []SortBy
	GetRootDirs() []string
	GetCustomRootDir() string
}

type Category Pair
type SortBy Pair

type Pair struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
}
