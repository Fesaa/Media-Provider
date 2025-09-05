package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

type TestRequestBody struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type InvalidTestBody struct {
	RequiredField string `json:"required_field" validate:"required"`
}

func TestWithBody(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupHandler   func() fiber.Handler
	}{
		{
			name: "Valid JSON body",
			requestBody: TestRequestBody{
				Name:  "John Doe",
				Age:   30,
				Email: "john@example.com",
			},
			expectedStatus: 200,
			setupHandler: func() fiber.Handler {
				return withBody(func(c *fiber.Ctx, body TestRequestBody) error {
					assert.Equal(t, "John Doe", body.Name)
					assert.Equal(t, 30, body.Age)
					assert.Equal(t, "john@example.com", body.Email)

					return c.JSON(fiber.Map{
						"message": "success",
						"data":    body,
					})
				})
			},
		},
		{
			name:           "Empty body with struct",
			requestBody:    TestRequestBody{},
			expectedStatus: 200,
			setupHandler: func() fiber.Handler {
				return withBody(func(c *fiber.Ctx, body TestRequestBody) error {
					assert.Empty(t, body.Name)
					assert.Empty(t, body.Age)
					assert.Empty(t, body.Email)

					return c.JSON(fiber.Map{"message": "empty body handled"})
				})
			},
		},
		{
			name:           "Invalid JSON body",
			requestBody:    `{"invalid": json}`, // Invalid JSON
			expectedStatus: 400,
			setupHandler: func() fiber.Handler {
				return withBody(func(c *fiber.Ctx, body TestRequestBody) error {
					t.Error("Handler should not be called with invalid JSON")
					return nil
				})
			},
		},
		{
			name:           "Malformed JSON",
			requestBody:    `{"name": "John", "age": "not_a_number"}`,
			expectedStatus: 400,
			setupHandler: func() fiber.Handler {
				return withBody(func(c *fiber.Ctx, body TestRequestBody) error {
					t.Error("Handler should not be called with malformed JSON")
					return nil
				})
			},
		},
		{
			name:           "Empty request body",
			requestBody:    nil,
			expectedStatus: 400,
			setupHandler: func() fiber.Handler {
				return withBody(func(c *fiber.Ctx, body TestRequestBody) error {
					t.Error("Handler should not be called with empty request body")
					return nil
				})
			},
		},
		{
			name:           "Invalid request body",
			requestBody:    "{}",
			expectedStatus: 400,
			setupHandler: func() fiber.Handler {
				return withBodyValidation(func(c *fiber.Ctx, body InvalidTestBody) error {
					t.Error("Handler should not be called with invalid request body")
					return nil
				})
			},
		},
		{
			name: "Valid request body",
			requestBody: TestRequestBody{
				Name:  "Amelia",
				Age:   21,
				Email: "amelia@localhost",
			},
			expectedStatus: 200,
			setupHandler: func() fiber.Handler {
				return withBodyValidation(func(c *fiber.Ctx, body TestRequestBody) error {
					assert.Equal(t, "Amelia", body.Name)
					assert.Equal(t, 21, body.Age)
					assert.Equal(t, "amelia@localhost", body.Email)
					return c.JSON(fiber.Map{})
				})
			},
		},
	}

	// It's fine to re-use
	validationService := services.ValidationServiceProvider(services.ValidatorProvider(), zerolog.Nop())
	container := dig.New()
	utils.Must(container.Provide(utils.Identity(validationService)))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(ctx *fiber.Ctx) error {
				ctx.Locals(services.ServiceProviderKey.Value(), container)
				ctx.Locals(services.LoggerKey.Value(), zerolog.Nop())
				return ctx.Next()
			}).Post("/test", tt.setupHandler())

			var reqBody []byte
			var err error

			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, err = json.Marshal(tt.requestBody)
					require.NoError(t, err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 400 {
				var errorResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)

				assert.Contains(t, errorResp, "error")
				assert.IsType(t, "", errorResp["error"])
			}
		})
	}
}

func TestWithBody_HandlerError(t *testing.T) {
	app := fiber.New()

	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Locals(services.LoggerKey.Value(), zerolog.Nop())
		return ctx.Next()
	})

	app.Post("/error", withBody(func(c *fiber.Ctx, body TestRequestBody) error {
		return fiber.NewError(500, "handler error")
	}))

	reqBody := TestRequestBody{Name: "John", Age: 30}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/error", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 500, resp.StatusCode)
}

func TestWithParam(t *testing.T) {
	tests := []struct {
		name           string
		setupRoute     func(*fiber.App)
		url            string
		expectedStatus int
		expectedValue  any
	}{
		{
			name: "Valid query parameter uint",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[uint]("id"), func(c *fiber.Ctx, id uint) error {
					assert.Equal(t, uint(123), id)
					return c.JSON(fiber.Map{"id": id})
				}))
			},
			url:            "/test?id=123",
			expectedStatus: 200,
			expectedValue:  uint(123),
		},
		{
			name: "Valid path parameter uint",
			setupRoute: func(app *fiber.App) {
				app.Get("/test/:id", withParam(newPathParam[uint]("id"), func(c *fiber.Ctx, id uint) error {
					assert.Equal(t, uint(456), id)
					return c.JSON(fiber.Map{"id": id})
				}))
			},
			url:            "/test/456",
			expectedStatus: 200,
			expectedValue:  uint(456),
		},
		{
			name: "Valid query parameter string",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[string]("name"), func(c *fiber.Ctx, name string) error {
					assert.Equal(t, "john", name)
					return c.JSON(fiber.Map{"name": name})
				}))
			},
			url:            "/test?name=john",
			expectedStatus: 200,
			expectedValue:  "john",
		},
		{
			name: "Valid query parameter int",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[int]("age"), func(c *fiber.Ctx, age int) error {
					assert.Equal(t, 25, age)
					return c.JSON(fiber.Map{"age": age})
				}))
			},
			url:            "/test?age=25",
			expectedStatus: 200,
			expectedValue:  25,
		},
		{
			name: "Valid query parameter bool",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[bool]("active"), func(c *fiber.Ctx, active bool) error {
					assert.True(t, active)
					return c.JSON(fiber.Map{"active": active})
				}))
			},
			url:            "/test?active=true",
			expectedStatus: 200,
			expectedValue:  true,
		},
		{
			name: "Missing required parameter",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[uint]("id"), func(c *fiber.Ctx, id uint) error {
					t.Error("Handler should not be called with missing parameter")
					return nil
				}))
			},
			url:            "/test",
			expectedStatus: 400,
		},
		{
			name: "Invalid parameter format",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[uint]("id"), func(c *fiber.Ctx, id uint) error {
					t.Error("Handler should not be called with invalid parameter")
					return nil
				}))
			},
			url:            "/test?id=invalid",
			expectedStatus: 400,
		},
		{
			name: "Parameter with default value - empty",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[uint]("id", withAllowEmpty[uint](999)), func(c *fiber.Ctx, id uint) error {
					assert.Equal(t, uint(999), id)
					return c.JSON(fiber.Map{"id": id})
				}))
			},
			url:            "/test",
			expectedStatus: 200,
			expectedValue:  uint(999),
		},
		{
			name: "Parameter with default value - provided",
			setupRoute: func(app *fiber.App) {
				app.Get("/test", withParam(newQueryParam[uint]("id", withAllowEmpty[uint](999)), func(c *fiber.Ctx, id uint) error {
					assert.Equal(t, uint(123), id)
					return c.JSON(fiber.Map{"id": id})
				}))
			},
			url:            "/test?id=123",
			expectedStatus: 200,
			expectedValue:  uint(123),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			tt.setupRoute(app)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 && tt.expectedValue != nil {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				var expectedForJSON interface{}
				switch v := tt.expectedValue.(type) {
				case uint:
					expectedForJSON = float64(v) // JSON numbers become float64
				case int:
					expectedForJSON = float64(v)
				default:
					expectedForJSON = v
				}

				for _, value := range response {
					assert.Equal(t, expectedForJSON, value)
					break
				}
			}

			if tt.expectedStatus == 400 {
				var errorResp map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp, "error")
			}
		})
	}
}
