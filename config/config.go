package config

import "log/slog"

type Config struct {
	SyncId   int    `json:"sync_id"`
	Port     string `json:"port"`
	Password string `json:"password"`
	RootDir  string `json:"root_dir"`
	BaseUrl  string `json:"base_url"`

	Logging    Logging    `json:"logging"`
	Downloader Downloader `json:"downloader"`

	Pages []Page `json:"pages"`
}

type Downloader struct {
	MaxConcurrentTorrents       int `json:"max_torrents"`
	MaxConcurrentMangadexImages int `json:"max_mangadex_images"`
}

type Logging struct {
	Level   slog.Level
	Source  bool
	Handler string
	LogHttp bool
}

type Page struct {
	Title         string              `json:"title"`
	Provider      []Provider          `json:"provider"`
	Modifiers     map[string]Modifier `json:"modifiers"`
	Dirs          []string            `json:"dirs"`
	CustomRootDir string              `json:"custom_root_dir"`
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

type Modifier struct {
	Title  string            `yaml:"title" json:"title"`
	Type   ModifierType      `yaml:"type" json:"type"`
	Values map[string]string `yaml:"values" json:"values"`
}
