package services

import (
	"context"
	"github.com/Fesaa/Media-Provider/api/auth"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/philippseith/signalr"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type SignalRService interface {
	signalr.HubInterface

	Broadcast(eventType payload.EventType, data interface{})

	SizeUpdate(id string, size string)
	ProgressUpdate(data payload.ContentProgressUpdate)
	StateUpdate(id string, state payload.ContentState)

	AddContent(data payload.InfoStat)
	UpdateContentInfo(data payload.InfoStat)
	DeleteContent(id string)

	// Notify may be used directly by anyone to send a quick toast to the frontend.
	// Use NotificationService for notification that must persist
	Notify(notification models.Notification)
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

	connectionHappened bool
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

func (s *signalrService) OnConnected(string) {
	s.connectionHappened = true
}

func (s *signalrService) Broadcast(eventType payload.EventType, data interface{}) {
	if !s.connectionHappened {
		s.log.Debug().Any("type", eventType).
			Msg("broadcasted notification won't be send out, as no connections have been made yet")
		return
	}
	s.Clients().All().Send(string(eventType), data)
}

func (s *signalrService) SizeUpdate(id string, size string) {
	s.Broadcast(payload.EventTypeContentSizeUpdate, payload.ContentSizeUpdate{
		ContentId: id,
		Size:      size,
	})
}

func (s *signalrService) ProgressUpdate(data payload.ContentProgressUpdate) {
	s.Broadcast(payload.EventTypeContentProgressUpdate, data)
}

func (s *signalrService) StateUpdate(id string, state payload.ContentState) {
	s.Broadcast(payload.EventTypeContentStateUpdate, payload.ContentStateUpdate{
		ContentId:    id,
		ContentState: state,
	})
}

func (s *signalrService) AddContent(data payload.InfoStat) {
	s.Broadcast(payload.EventTypeAddContent, data)
}

func (s *signalrService) UpdateContentInfo(data payload.InfoStat) {
	s.Broadcast(payload.EventTypeContentInfoUpdate, data)
}

func (s *signalrService) DeleteContent(id string) {
	s.Broadcast(payload.EventTypeDeleteContent, payload.DeleteContent{ContentId: id})
}

func (s *signalrService) Notify(notification models.Notification) {
	s.Broadcast(payload.EventTypeNotification, notification)
}

func (s *signalrService) setup(app *fiber.App) error {
	server, err := signalr.NewServer(context.TODO(), signalr.UseHub(s),
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
