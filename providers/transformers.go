package providers

import (
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/irevenko/go-nyaa/nyaa"
	"net/url"
)

func subsPleaseTransformer(s SearchRequest) subsplease.SearchOptions {
	return subsplease.SearchOptions{
		Query: s.Query,
	}
}

func limeTransformer(s SearchRequest) limetorrents.SearchOptions {
	return limetorrents.SearchOptions{
		Category: limetorrents.ConvertCategory(s.Category),
		Query:    s.Query,
		Page:     1,
	}
}

func nyaaTransformer(s SearchRequest) nyaa.SearchOptions {
	n := nyaa.SearchOptions{}
	n.Query = url.QueryEscape(s.Query)
	if s.Provider != "" {
		n.Provider = string(s.Provider)
	} else {
		n.Provider = "nyaa"
	}

	if s.Category != "" {
		n.Category = s.Category
	}

	if s.SortBy != "" {
		n.SortBy = s.SortBy
	}

	if s.Filter != "" {
		n.Filter = s.Filter
	}

	return n
}

func ytsTransformer(s SearchRequest) yts.SearchOptions {
	y := yts.SearchOptions{}
	y.Query = s.Query
	y.SortBy = s.SortBy
	y.Page = 1
	return y
}
