package services

import (
	"context"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/philippseith/signalr"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type SignalRService interface {
	signalr.HubInterface
}

type SignalRParams struct {
	dig.In
	Log  zerolog.Logger
	Auth auth.Provider `name:"jwt-auth"`
}

type signalrService struct {
	signalr.Hub

	app    *fiber.App
	server signalr.Server
	auth   auth.Provider
	log    zerolog.Logger
}

func SignalRServiceProvider(params SignalRParams) SignalRService {
	return &signalrService{
		auth: params.Auth,
		log:  params.Log.With().Str("handler", "signalR-service").Logger(),
	}
}

func RegisterSignalREndPoint(service SignalRService, app *fiber.App) error {
	return (service.(*signalrService)).setup(app)
}

func (s *signalrService) setup(app *fiber.App) error {
	server, err := signalr.NewServer(context.TODO(), signalr.SimpleHubFactory(s),
		signalr.Logger(&kitLoggerAdapter{log: s.log}, false))
	if err != nil {
		return err
	}

	s.server = server
	s.app = app

	server.MapHTTP(func() signalr.MappableRouter {
		return s
	}, "/ws")

	return nil
}
