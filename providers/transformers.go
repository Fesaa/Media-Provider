package providers

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/webtoon"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/irevenko/go-nyaa/nyaa"
	"net/url"
)

func mangadexTransformer(s payload.SearchRequest) mangadex.SearchOptions {
	ms := mangadex.SearchOptions{
		Query:            s.Query,
		SkipNotFoundTags: true,
	}

	iT, ok := s.Modifiers["includeTags"]
	if ok {
		ms.IncludedTags = iT
	}
	eT, ok := s.Modifiers["excludeTags"]
	if ok {
		ms.ExcludedTags = eT
	}
	st, ok := s.Modifiers["status"]
	if ok {
		ms.Status = st
	}
	cr, ok := s.Modifiers["contentRating"]
	if ok {
		ms.ContentRating = cr
	}
	pd, ok := s.Modifiers["publicationDemographic"]
	if ok {
		ms.PublicationDemographic = pd
	}

	return ms
}

func subsPleaseTransformer(s payload.SearchRequest) subsplease.SearchOptions {
	return subsplease.SearchOptions{
		Query: s.Query,
	}
}

func limeTransformer(s payload.SearchRequest) limetorrents.SearchOptions {
	categories, ok := s.Modifiers["categories"]
	var category string
	if ok && len(categories) > 0 {
		category = categories[0]
	}
	return limetorrents.SearchOptions{
		Category: limetorrents.ConvertCategory(category),
		Query:    s.Query,
		Page:     1,
	}
}

func nyaaTransformer(p config.Provider) requestTransformerFunc[nyaa.SearchOptions] {
	var ps string
	switch p {
	case config.NYAA:
		ps = "nyaa"
		break
	case config.SUKEBEI:
		ps = "sukebei"
		break
	default:
		panic("Invalid provider")
	}
	return func(s payload.SearchRequest) nyaa.SearchOptions {
		n := nyaa.SearchOptions{}
		n.Query = url.QueryEscape(s.Query)
		n.Provider = ps
		categories, ok := s.Modifiers["categories"]
		if ok && len(categories) > 0 {
			n.Category = categories[0]
		}

		sortBys, ok := s.Modifiers["sortBys"]
		if ok && len(sortBys) > 0 {
			n.SortBy = sortBys[0]
		}

		filters, ok := s.Modifiers["filters"]
		if ok && len(filters) > 0 {
			n.Filter = filters[0]
		}

		return n
	}
}

func ytsTransformer(s payload.SearchRequest) yts.SearchOptions {
	y := yts.SearchOptions{}
	y.Query = s.Query
	sortBys, ok := s.Modifiers["sortBys"]
	if ok && len(sortBys) > 0 {
		y.SortBy = sortBys[0]
	}
	y.Page = 1
	return y
}

func webtoonTransformer(s payload.SearchRequest) webtoon.SearchOptions {
	return webtoon.SearchOptions{
		Query: s.Query,
	}
}
