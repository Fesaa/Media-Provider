package services

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
	"testing"
)

var service = ValidationServiceProvider(ValidatorProvider(), zerolog.New(zerolog.ConsoleWriter{}))

func TestSearchRequest(t *testing.T) {
	sr := payload.SearchRequest{
		Provider:  []models.Provider{models.MANGADEX},
		Query:     "",
		Modifiers: nil,
	}
	if err := service.Validate(sr); err != nil {
		t.Error(err)
	}

	sr = payload.SearchRequest{
		Provider:  []models.Provider{models.Provider(9999)},
		Query:     "",
		Modifiers: nil,
	}

	if err := service.Validate(sr); err == nil {
		t.Error("Expected error, as provider is invalid")
	}

	sr = payload.SearchRequest{
		Provider: []models.Provider{models.MANGADEX},
		Query:    "",
		Modifiers: map[string][]string{
			"foo": {"bar"},
			"bar": {"foo"},
		},
	}
	if err := service.Validate(sr); err != nil {
		t.Error(err)
	}

	sr = payload.SearchRequest{
		Provider: []models.Provider{models.MANGADEX},
		Query:    "",
		Modifiers: map[string][]string{
			"foo": {"bar"},
			"bar": {"foo"},
			"":    {"abc"},
		},
	}

	if err := service.Validate(sr); err == nil {
		t.Error("Expected error, as Modifiers key is invalid")
	}
}

func TestProvider(t *testing.T) {
	type testStruct struct {
		Provider models.Provider `validate:"provider"`
	}

	if err := service.Validate(&testStruct{Provider: models.Provider(9999)}); err == nil {
		t.Error("Expected error, as provider is invalid")
	}

	if err := service.Validate(&testStruct{Provider: models.MANGADEX}); err != nil {
		t.Error(err)
	}
}

func TestDiff(t *testing.T) {
	type testStruct struct {
		One string
		Two string `validate:"diff=One"`
	}

	if err := service.Validate(&testStruct{
		One: "one",
		Two: "two",
	}); err != nil {
		t.Error(err)
	}

	if err := service.Validate(&testStruct{
		One: "one",
		Two: "one",
	}); err == nil {
		t.Error("Expected error, as One is different")
	}
}

func TestSwapPage(t *testing.T) {
	r := payload.SwapPageRequest{
		Id1: 0,
		Id2: 0,
	}

	if err := service.Validate(&r); err == nil {
		t.Error("Expected error, as id2 is the same ad id1")
	}

	r = payload.SwapPageRequest{
		Id1: 1,
		Id2: 2,
	}

	if err := service.Validate(&r); err != nil {
		t.Error(err)
	}
}
