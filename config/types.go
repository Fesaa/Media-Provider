package config

// TODO: Validation
type Config struct {
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	RootDir  string `yaml:"root_dir"`
	RootURL  string `yaml:"root_url"`
	Pages    []Page `yaml:"pages"`
}

type Page struct {
	Title        string       `yaml:"title" json:"title"`
	SearchConfig SearchConfig `yaml:"search" json:"search"`
}

type SearchProvider string

const (
	NYAA SearchProvider = "nyaa"
	YTS  SearchProvider = "yts"
	LIME SearchProvider = "lime"
)

type SearchConfig struct {
	Provider      SearchProvider `yaml:"provider" json:"provider"`
	Categories    []Category     `yaml:"categories" json:"categories"`
	SortBys       []SortBy       `yaml:"sorts" json:"sorts"`
	RootDirs      []string       `yaml:"root_dirs" json:"root_dirs"`
	CustomRootDir string         `yaml:"custom_root_dir" json:"custom_root_dir"`
}

type Category Pair
type SortBy Pair

type Pair struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
}
