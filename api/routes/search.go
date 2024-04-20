package routes

import (
	"log/slog"
	"net/url"

	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/gofiber/fiber/v2"
	"github.com/irevenko/go-nyaa/nyaa"
)

type SearchRequest struct {
	Provider string `json:"provider,omitempty"`
	Query    string `json:"query"`
	Category string `json:"category,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	Filter   string `json:"filter,omitempty"`
}

func Search(ctx *fiber.Ctx) error {
	var searchRequest SearchRequest
	if err := ctx.BodyParser(&searchRequest); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	switch searchRequest.Provider {
	case "nyaa":
		return nyaaSearch(ctx, searchRequest)
	case "yts":
		return ytsSearch(ctx, searchRequest)
	case "limetorrents":
		return limeTorrensSearch(ctx, searchRequest)
	default:
		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: "Invalid provider, can't process request.",
		}
	}
}

func limeTorrensSearch(ctx *fiber.Ctx, r SearchRequest) error {
	limeS := r.ToLimeTorrents()
	torrents, err := limetorrents.Search(limeS)
	if err != nil {
		slog.Error("Error searching limetorrents", "err", err)
		return err
	}

	slog.Info("Found torrents", "amount", len(torrents))
	return ctx.JSON(fromLime(torrents))
}

func ytsSearch(ctx *fiber.Ctx, r SearchRequest) error {
	ytsS := r.ToYTS()
	req, err := yts.Search(ytsS)
	if err != nil {
		slog.Error("Error searching yts", "err", err)
		return err
	}
	slog.Info("Found movies", "amount", len(req.Data.Movies))
	return ctx.JSON(fromYTS(req.Data.Movies))
}

func nyaaSearch(ctx *fiber.Ctx, r SearchRequest) error {
	nyaaS := r.ToNyaa()
	torrents, err := nyaa.Search(nyaaS)
	if err != nil {
		slog.Error("Error searching nyaa", "err", err)
		return err
	}
	slog.Info("Found torrents", "amount", len(torrents))
	return ctx.JSON(torrents)
}

func (s *SearchRequest) ToNyaa() nyaa.SearchOptions {
	n := nyaa.SearchOptions{}
	n.Query = url.QueryEscape(s.Query)
	if s.Provider != "" {
		n.Provider = s.Provider
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

func (s *SearchRequest) ToYTS() yts.YTSSearchOptions {
	y := yts.YTSSearchOptions{}
	y.Query = s.Query
	y.SortBy = s.SortBy
	y.Page = 1
	return y
}

func (s *SearchRequest) ToLimeTorrents() limetorrents.SearchOptions {
	return limetorrents.SearchOptions{
		Category: limetorrents.ConvertCategory(s.Category),
		Query:    s.Query,
		Page:     1,
	}
}
