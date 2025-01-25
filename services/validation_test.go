package services

import (
	"bytes"
	"encoding/json"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
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

func TestDiffInvalid(t *testing.T) {
	type testStruct struct {
		One string
		Two string `validate:"diff=Three"`
	}

	if err := service.Validate(testStruct{
		One: "",
		Two: "",
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

func TestValidationService_ValidateCtxBadBody(t *testing.T) {
	app := fiber.New()

	app.Post("/", func(c *fiber.Ctx) error {
		var login payload.LoginRequest
		if err := service.ValidateCtx(c, &login); err == nil {
			t.Error("Should have error")
		}
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte{1, 4, 2, 9}))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if _, err := app.Test(req, -1); err != nil {
		t.Error(err)
	}

}

func TestValidationService_ValidateCtxInvalidStruct(t *testing.T) {
	app := fiber.New()

	app.Post("/", func(c *fiber.Ctx) error {
		var login payload.LoginRequest
		if err := service.ValidateCtx(c, &login); err != nil {
			t.Error(err)
		}
		return nil
	})

	logingRequest := payload.LoginRequest{
		UserName: "username",
		Password: "password",
		Remember: false,
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(&logingRequest)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	if _, err = app.Test(req, -1); err != nil {
		t.Error(err)
	}
}
