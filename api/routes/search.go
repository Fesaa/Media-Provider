package routes

import (
	"net/url"

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

	nyaaSearch := searchRequest.ToNyaa()

	torrents, err := nyaa.Search(nyaaSearch)
	if err != nil {
		return err
	}

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
