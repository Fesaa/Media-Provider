package mangadex

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"net/http"
	"strconv"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *http.Client
	ps         api.Client
	repository Repository
}

func (b *Builder) Provider() models.Provider {
	return models.MANGADEX
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(mangas *MangaSearchResponse) []payload.Info {
	if mangas == nil {
		return []payload.Info{}
	}

	info := make([]payload.Info, 0)
	for _, data := range mangas.Data {
		enTitle := data.Attributes.LangTitle("en")
		info = append(info, payload.Info{
			Name:        enTitle,
			Description: data.Attributes.LangDescription("en"),
			Size: func() string {
				s := ""
				if data.Attributes.LastVolume != "" {
					s += fmt.Sprintf("%s Vol.", data.Attributes.LastVolume)
				}

				if data.Attributes.LastChapter != "" {
					s += fmt.Sprintf(" %s Ch.", data.Attributes.LastChapter)
				}
				return s
			}(),
			Tags: []payload.InfoTag{
				payload.Of("Date", strconv.Itoa(data.Attributes.Year)),
			},
			InfoHash: data.Id,
			RefUrl:   data.RefURL(),
			Provider: models.MANGADEX,
			ImageUrl: data.CoverURL(),
		})
	}

	return info
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	ms := SearchOptions{
		Query: s.Query,
	}

	skip, ok := s.Modifiers["SkipNotFoundTags"]
	if ok {
		ms.SkipNotFoundTags = utils.OrDefault(skip, "true") == "true"
	} else {
		ms.SkipNotFoundTags = true
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

func (b *Builder) Search(s SearchOptions) (*MangaSearchResponse, error) {
	return b.repository.SearchManga(context.TODO(), s)
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{
		Definitions: []payload.DownloadMetadataDefinition{
			{
				Key:           LanguageKey,
				FormType:      payload.DROPDOWN,
				DefaultOption: "en",
				Options:       languages,
			},
			{
				Key:      ScanlationGroupKey,
				FormType: payload.TEXT,
			},
			{
				Key:      DownloadOneShotKey,
				FormType: payload.SWITCH,
			},
			{
				Key:           IncludeCover,
				FormType:      payload.SWITCH,
				DefaultOption: "true",
			},
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ps
}

func NewBuilder(log zerolog.Logger, httpClient *http.Client, ps api.Client, repository Repository) *Builder {
	return &Builder{log.With().Str("handler", "mangadex-provider").Logger(),
		httpClient, ps, repository}
}
