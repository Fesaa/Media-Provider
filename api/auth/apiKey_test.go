package auth

import (
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAuth struct {
	succes bool
}

func (m *mockAuth) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	if m.succes {
		return true, nil
	}
	return false, nil
}

func (m *mockAuth) Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	return nil, nil
}

func (m *mockAuth) Middleware(ctx *fiber.Ctx) error {
	b, err := m.IsAuthenticated(ctx)
	if err != nil {
		return err
	}
	if !b {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Next()
}

type mockUsers struct {
	user      *models.User
	apiKey    string
	getById   func(id uint) (*models.User, error)
	getByName func(name string) (*models.User, error)
}

func (m mockUsers) All() ([]models.User, error) {
	panic("implement me")
}

func (m mockUsers) ExistsAny() (bool, error) {
	panic("implement me")
}

func (m mockUsers) GetById(id uint) (*models.User, error) {
	if m.getById != nil {
		return m.getById(id)
	}
	return m.user, nil
}

func (m mockUsers) GetByName(name string) (*models.User, error) {
	if m.getByName != nil {
		return m.getByName(name)
	}
	return m.user, nil
}

func (m mockUsers) GetByApiKey(key string) (*models.User, error) {
	if m.apiKey == key {
		return &models.User{
			ApiKey: m.apiKey,
		}, nil
	}
	return nil, nil
}

func (m mockUsers) Create(name string, opts ...models.Option[models.User]) (*models.User, error) {
	panic("implement me")
}

func (m mockUsers) Update(user models.User, opts ...models.Option[models.User]) (*models.User, error) {
	panic("implement me")
}

func (m mockUsers) UpdateById(id uint, opts ...models.Option[models.User]) (*models.User, error) {
	panic("implement me")
}

func (m mockUsers) GenerateReset(userId uint) (*models.PasswordReset, error) {
	panic("implement me")
}

func (m mockUsers) GetResetByUserId(userId uint) (*models.PasswordReset, error) {
	panic("implement me")
}

func (m mockUsers) GetReset(key string) (*models.PasswordReset, error) {
	panic("implement me")
}

func (m mockUsers) DeleteReset(key string) error {
	panic("implement me")
}

func (m mockUsers) Delete(id uint) error {
	panic("implement me")
}

func TestApiKeyAuth_IsAuthenticated(t *testing.T) {
	jwt := mockAuth{false}
	auth := NewApiKeyAuth(apiKeyAuthParams{
		DB: &db.Database{Users: mockUsers{
			apiKey: "test",
		}},
		JWT: &jwt,
		Log: zerolog.New(io.Discard),
	})

	app := fiber.New()

	app.Use(auth.Middleware)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{})
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Errorf("TestApiKeyAuth_IsAuthenticated() error = %v, resp %v", err, resp)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("TestApiKeyAuth_IsAuthenticated() error = %v, resp %v", resp.StatusCode, resp)
	}

	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/?api-key=test", nil))
	if err != nil {
		t.Errorf("TestApiKeyAuth_IsAuthenticated() error = %v, resp %v", err, resp)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("TestApiKeyAuth_IsAuthenticated() error = %v, resp %v", resp.StatusCode, resp)
	}

	jwt.succes = true
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Errorf("TestApiKeyAuth_IsAuthenticated() error = %v, resp %v", err, resp)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("TestApiKeyAuth_IsAuthenticated() error = %v, resp %v", resp.StatusCode, resp)
	}
}

func TestApiKeyAuth_Login(t *testing.T) {
	_, err := (apiKeyAuth{}).Login(payload.LoginRequest{})
	if err == nil {
		t.Errorf("should have errored")
	}
}
