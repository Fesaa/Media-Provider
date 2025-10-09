package dynasty

import (
	"strings"

	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
)

type SearchOptions struct {
	Query string
}

type SearchData struct {
	Id      string
	Title   string
	Authors []publication.Person
	Tags    []publication.Tag
}

func (s *SearchData) RefUrl() string {
	if strings.HasPrefix(s.Id, "/series/") {
		return DOMAIN + s.Id
	}
	return DOMAIN + "/series/" + s.Id
}

type SeriesStatus string

const (
	Completed SeriesStatus = "Completed"
	Ongoing   SeriesStatus = "Ongoing"
)

func toPublicationStatus(status string) publication.Status {
	switch status {
	case "Completed":
		return publication.StatusCompleted
	case "Ongoing":
		return publication.StatusOngoing
	}

	return ""
}
