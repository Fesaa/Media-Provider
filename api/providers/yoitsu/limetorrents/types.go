package limetorrents

import "strings"

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

func ConvertCategory(c string) Category {
	switch strings.ToLower(c) {
	case "anime":
		return ANIME
	case "applications":
		return APPS
	case "games":
		return GAMES
	case "movies":
		return MOVIE
	case "music":
		return MUSIC
	case "tv":
		return TV
	case "other":
		return OTHER
	default:
		return ALL
	}
}

type SearchResult struct {
	Name    string
	Url     string
	Hash    string
	Size    string
	Seed    string
	Leach   string
	Added   string
	PageUrl string
}

type SearchOptions struct {
	Category Category
	Query    string
	Page     int
}
