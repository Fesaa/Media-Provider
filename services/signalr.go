package services

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/philippseith/signalr"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

const (
	allDownloadInfoGroup = "AllDownloadInfoGroup"
)

type SignalRService interface {
	signalr.HubInterface

	Broadcast(eventType payload.EventType, data interface{})

	SizeUpdate(uint, string, string)
	ProgressUpdate(uint, payload.ContentProgressUpdate)
	StateUpdate(uint, string, payload.ContentState)

	AddContent(uint, payload.InfoStat)
	UpdateContentInfo(uint, payload.InfoStat)
	DeleteContent(string)

	// Notify may be used directly by anyone to send a quick toast to the frontend.
	// Use NotificationService for notification that must persist
	Notify(notification models.Notification)
}

type SignalRParams struct {
	dig.In
	Log  zerolog.Logger
	Auth AuthService
}

type signalrService struct {
	signalr.Hub

	app    *fiber.App
	server signalr.Server
	auth   AuthService
	log    zerolog.Logger

	connectionHappened bool
	clients            utils.SafeMap[uint, string]
}

func SignalRServiceProvider(params SignalRParams) SignalRService {
	return &signalrService{
		auth:    params.Auth,
		clients: utils.NewSafeMap[uint, string](),
		log:     params.Log.With().Str("handler", "signalR-service").Logger(),
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

func (s *signalrService) sendToUserAndGroup(userId uint, group string, eventType payload.EventType, data interface{}) {
	if !s.connectionHappened {
		return
	}

	clientId, ok := s.clients.Get(userId)
	s.log.Debug().Str("id", clientId).Str("group", group).Str("eventType", string(eventType)).
		Msg("sending to user")
	if ok {
		s.Clients().Client(clientId).Send(string(eventType), data)
	}

	s.Clients().Group(group).Send(string(eventType), data)
}

func (s *signalrService) SizeUpdate(userId uint, id string, size string) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentSizeUpdate, payload.ContentSizeUpdate{
		ContentId: id,
		Size:      size,
	})
}

func (s *signalrService) ProgressUpdate(userId uint, data payload.ContentProgressUpdate) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentProgressUpdate, data)
}

func (s *signalrService) StateUpdate(userId uint, id string, state payload.ContentState) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentStateUpdate, payload.ContentStateUpdate{
		ContentId:    id,
		ContentState: state,
	})
}

func (s *signalrService) AddContent(userId uint, data payload.InfoStat) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeAddContent, data)
}

func (s *signalrService) UpdateContentInfo(userId uint, data payload.InfoStat) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentInfoUpdate, data)
}

func (s *signalrService) DeleteContent(id string) {
	s.Broadcast(payload.EventTypeDeleteContent, payload.DeleteContent{ContentId: id})
}

func (s *signalrService) Notify(notification models.Notification) {
	s.Broadcast(payload.EventTypeNotification, notification)
}

func (s *signalrService) setup(app *fiber.App) error {
	server, err := signalr.NewServer(context.Background(), signalr.UseHub(s),
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
