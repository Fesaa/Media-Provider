package dynasty

import (
	"strings"
	"time"
)

type SearchOptions struct {
	Query string
}

type SearchData struct {
	Id      string
	Title   string
	Authors []Author
	Tags    []Tag
}

type Author Identifiable

type Tag Identifiable

type Identifiable struct {
	DisplayName string
	Id          string
}

func (s *SearchData) RefUrl() string {
	if strings.HasPrefix(s.Id, "/series/") {
		return DOMAIN + s.Id
	}
	return DOMAIN + "/series/" + s.Id
}

type Series struct {
	Id          string
	Title       string
	AltTitle    string
	Description string
	Status      SeriesStatus
	CoverUrl    string

	Authors []Author
	Tags    []Tag

	Chapters []Chapter
}

func (s *Series) RefUrl() string {
	return seriesURL(s.Id)
}

type Chapter struct {
	Id          string
	Title       string
	Volume      string
	Chapter     string
	ReleaseDate *time.Time
	Tags        []Tag
	Authors     []Author
}

type SeriesStatus string

const (
	Completed SeriesStatus = "Completed"
	Ongoing   SeriesStatus = "Ongoing"
)
