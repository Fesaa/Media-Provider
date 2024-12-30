package providers

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	dynasty_scans "github.com/Fesaa/Media-Provider/providers/pasloe/dynasty-scans"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
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

func nyaaTransformer(p models.Provider) requestTransformerFunc[nyaa.SearchOptions] {
	var ps string
	switch p {
	case models.NYAA:
		ps = "nyaa"
		break
	case models.SUKEBEI:
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

func webtoonTransformer(s payload.SearchRequest) webtoon.SearchOptions {
	return webtoon.SearchOptions{
		Query: s.Query,
	}
}

func dynastyTransformer(s payload.SearchRequest) dynasty_scans.SearchOptions {
	return dynasty_scans.SearchOptions{
		Query: s.Query,
	}
}
