package limetorrents

type Category string

const (
	ALL   Category = "all"
	ANIME          = "anime"
	APPS           = "applications"
	GAMES          = "games"
	MOVIE          = "movies"
	MUSIC          = "music"
	TV             = "tv"
	OTHER          = "other"
)

type SearchResult struct {
	Name  string
	Url   string
	Size  string
	Seed  string
	Leach string
	Added string
}

type SearchOptions struct {
	Category Category
	Query    string
	Page     int
}
