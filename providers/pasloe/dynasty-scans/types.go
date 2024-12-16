package dynasty_scans

import "strings"

type SearchOptions struct {
	Query string
}

type SearchData struct {
	Id      string
	Title   string
	Authors []string
	Tags    []string
}

func (s *SearchData) RefUrl() string {
	if strings.HasPrefix(s.Id, "/series/") {
		return DOMAIN + s.Id
	}
	return DOMAIN + "/series/" + s.Id
}
