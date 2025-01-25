package auth

import (
	"encoding/base64"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware_Success(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{Secret: "testsecret"}
	log := zerolog.Nop()
	mockDB := &db.Database{
		Users: mockUsers{
			user: &models.User{
				Model: gorm.Model{
					ID: 1,
				},
				Name:   "testuser",
				ApiKey: "testapikey",
			},
		},
	}
	authHandler := NewJwtAuth(mockDB, cfg, log)

	app.Use(authHandler.Middleware)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{})
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	})
	validToken, _ := token.SignedString([]byte(cfg.Secret))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("got status code %d wanted 200", resp.StatusCode)
	}
}

func TestMiddleware_Unauthorized(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{Secret: "testsecret"}
	log := zerolog.Nop()
	mockDB := &db.Database{
		Users: mockUsers{},
	}
	authHandler := NewJwtAuth(mockDB, cfg, log)

	app.Use(authHandler.Middleware)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")

	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("got status code %d wanted 401", resp.StatusCode)
	}
}

func TestLogin_Success(t *testing.T) {
	cfg := &config.Config{Secret: "testsecret"}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	log := zerolog.Nop()
	mockDB := &db.Database{
		Users: mockUsers{
			user: &models.User{
				Model: gorm.Model{
					ID: 1,
				},
				Name:         "testuser",
				PasswordHash: base64.StdEncoding.EncodeToString(passwordBytes),
			},
			getByName: func(name string) (*models.User, error) {
				return &models.User{
					Model: gorm.Model{
						ID: 1,
					},
					Name:         "testuser",
					PasswordHash: base64.StdEncoding.EncodeToString(passwordBytes),
				}, nil
			},
		},
	}

	authHandler := NewJwtAuth(mockDB, cfg, log)

	loginRequest := payload.LoginRequest{
		UserName: "testuser",
		Password: "password",
	}

	loginResponse, err := authHandler.Login(loginRequest)
	if err != nil {
		t.Error(err)
	}

	if loginResponse == nil {
		t.Fatal("login response is nil")
	}

	if loginResponse.Name != "testuser" {
		t.Error("login response name is wrong")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	cfg := &config.Config{Secret: "testsecret"}
	log := zerolog.Nop()
	mockDB := &db.Database{
		Users: mockUsers{
			getByName: func(name string) (*models.User, error) {
				return &models.User{
					Model: gorm.Model{
						ID: 1,
					},
					Name:         "testuser",
					PasswordHash: "$2a$10$6Pz8lk.k3mr5Y/IFWyKNzuqs4eNFswhKoRJQ./XhN44Qo1HjFs.Ga",
				}, nil
			},
		},
	}

	authHandler := NewJwtAuth(mockDB, cfg, log)

	loginRequest := payload.LoginRequest{
		UserName: "testuser",
		Password: "wrongpassword",
	}

	loginResponse, err := authHandler.Login(loginRequest)
	if err == nil {
		t.Error("Expected an error")
	}

	if loginResponse != nil {
		t.Error("Expected nil response")
	}
}
