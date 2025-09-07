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
	allDownloadInfoGroup = string(models.ViewAllDownloads)
)

type SignalRService interface {
	signalr.HubInterface

	Broadcast(eventType payload.EventType, data interface{})

	SizeUpdate(int, string, string)
	ProgressUpdate(int, payload.ContentProgressUpdate)
	StateUpdate(int, string, payload.ContentState)

	AddContent(int, payload.InfoStat)
	UpdateContentInfo(int, payload.InfoStat)
	DeleteContent(int, string)

	// Notify may be used directly by anyone to send a quick toast to the frontend.
	// Use NotificationService for notification that must persist
	Notify(context.Context, models.Notification)
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
	clients            utils.SafeMap[int, string]
}

func SignalRServiceProvider(params SignalRParams) SignalRService {
	return &signalrService{
		auth:    params.Auth,
		clients: utils.NewSafeMap[int, string](),
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

func (s *signalrService) sendToUserAndGroup(userId int, group string, eventType payload.EventType, data interface{}) {
	if !s.connectionHappened {
		return
	}

	clientId, ok := s.clients.Get(userId)
	if ok {
		s.Clients().Client(clientId).Send(string(eventType), data)
	}

	s.Clients().Group(group).Send(string(eventType), data)
}

func (s *signalrService) SizeUpdate(userId int, id string, size string) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentSizeUpdate, payload.ContentSizeUpdate{
		ContentId: id,
		Size:      size,
	})
}

func (s *signalrService) ProgressUpdate(userId int, data payload.ContentProgressUpdate) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentProgressUpdate, data)
}

func (s *signalrService) StateUpdate(userId int, id string, state payload.ContentState) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentStateUpdate, payload.ContentStateUpdate{
		ContentId:    id,
		ContentState: state,
	})
}

func (s *signalrService) AddContent(userId int, data payload.InfoStat) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeAddContent, data)
}

func (s *signalrService) UpdateContentInfo(userId int, data payload.InfoStat) {
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeContentInfoUpdate, data)
}

func (s *signalrService) DeleteContent(userId int, id string) {
	data := payload.DeleteContent{ContentId: id}
	s.sendToUserAndGroup(userId, allDownloadInfoGroup, payload.EventTypeDeleteContent, data)
}

func (s *signalrService) Notify(ctx context.Context, notification models.Notification) {
	if !s.connectionHappened {
		return
	}

	if notification.Owner.Valid {
		s.sendToUser(int(notification.Owner.Int32), notification)
		return
	}

	if len(notification.RequiredRoles) > 0 {
		for _, role := range notification.RequiredRoles {
			s.Clients().Group(role).Send(string(payload.EventTypeNotification), notification)
			s.Clients().Group(role).Send(string(payload.EvenTypeNotificationAdd), fiber.Map{})
		}
		return
	}

	s.Broadcast(payload.EventTypeNotification, notification)
	s.Broadcast(payload.EvenTypeNotificationAdd, fiber.Map{})
}

func (s *signalrService) sendToUser(userId int, n models.Notification) {
	connId, ok := s.clients.Get(userId)
	if !ok {
		return
	}

	s.Clients().Client(connId).Send(string(payload.EventTypeNotification), n)
	s.Clients().Group(connId).Send(string(payload.EvenTypeNotificationAdd), fiber.Map{})
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
