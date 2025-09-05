package mock

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/philippseith/signalr"
)

type SignalR struct {
	signalr.Hub
}

func (m *SignalR) UpdateContentInfo(userId uint, data payload.InfoStat) {
}

func (m *SignalR) Broadcast(eventType payload.EventType, data interface{}) {
}

func (m *SignalR) SizeUpdate(userId uint, id string, size string) {
}

func (m *SignalR) ProgressUpdate(userId uint, data payload.ContentProgressUpdate) {
}

func (m *SignalR) StateUpdate(userId uint, id string, state payload.ContentState) {
}

func (m *SignalR) AddContent(userId uint, data payload.InfoStat) {
}

func (m *SignalR) DeleteContent(id string) {
}

func (m *SignalR) Notify(models.Notification) {}
