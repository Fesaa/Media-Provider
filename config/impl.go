package config

import (
	"log/slog"
)

type configImpl struct {
	Port          string            `yaml:"port"`
	Password      string            `yaml:"password"`
	RootDir       string            `yaml:"root_dir"`
	RootURL       string            `yaml:"root_url"`
	Pages         []pageImpl        `yaml:"pages"`
	LoggingConfig loggingConfigImpl `yaml:"logging"`
}

func (c configImpl) GetPort() string {
	return c.Port
}

func (c configImpl) GetPassWord() string {
	return c.Password
}

func (c configImpl) GetRootDir() string {
	return c.RootDir
}

func (c configImpl) GetRootURl() string {
	return c.RootURL
}

func (c configImpl) GetPages() Pages {
	pages := make([]Page, len(c.Pages))
	for i, page := range c.Pages {
		pages[i] = page
	}
	return pages
}

func (c configImpl) GetLoggingConfig() LoggingConfig {
	return c.LoggingConfig
}

type loggingConfigImpl struct {
	LogLevel slog.Level `yaml:"log_level"`
	Source   bool       `yaml:"source"`
	Handler  string     `yaml:"handler"`
}

func (l loggingConfigImpl) GetLogLevel() slog.Level {
	return l.LogLevel
}

func (l loggingConfigImpl) GetSource() bool {
	return l.Source
}

func (l loggingConfigImpl) GetHandler() string {
	return l.Handler
}

type pageImpl struct {
	Title        string           `yaml:"title" json:"title"`
	SearchConfig searchConfigImpl `yaml:"search" json:"search"`
}

func (p pageImpl) GetTitle() string {
	return p.Title
}

func (p pageImpl) GetSearchConfig() SearchConfig {
	return p.SearchConfig
}

type searchConfigImpl struct {
	Provider      SearchProvider `yaml:"provider" json:"provider"`
	Categories    []Category     `yaml:"categories" json:"categories"`
	SortBys       []SortBy       `yaml:"sorts" json:"sorts"`
	RootDirs      []string       `yaml:"root_dirs" json:"root_dirs"`
	CustomRootDir string         `yaml:"custom_root_dir" json:"custom_root_dir"`
}

func (s searchConfigImpl) GetProvider() SearchProvider {
	return s.Provider
}

func (s searchConfigImpl) GetCategories() []Category {
	return s.Categories
}

func (s searchConfigImpl) GetSortBys() []SortBy {
	return s.SortBys
}

func (s searchConfigImpl) GetRootDirs() []string {
	return s.RootDirs
}

func (s searchConfigImpl) GetCustomRootDir() string {
	return s.CustomRootDir
}
