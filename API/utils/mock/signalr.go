package mock

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/philippseith/signalr"
)

type SignalR struct {
	signalr.Hub
}

func (s *SignalR) Broadcast(eventType payload.EventType, data interface{}) {
}

func (s *SignalR) SizeUpdate(i int, s3 string, s2 string) {
}

func (s *SignalR) ProgressUpdate(i int, update payload.ContentProgressUpdate) {
}

func (s *SignalR) StateUpdate(i int, s2 string, state payload.ContentState) {
}

func (s *SignalR) AddContent(i int, stat payload.InfoStat) {
}

func (s *SignalR) UpdateContentInfo(i int, stat payload.InfoStat) {
}

func (s *SignalR) DeleteContent(i int, s2 string) {
}

func (s *SignalR) Notify(ctx context.Context, notification models.Notification) {
}
