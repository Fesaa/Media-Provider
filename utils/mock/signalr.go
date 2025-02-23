package mock

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/philippseith/signalr"
)

type SignalR struct {
	signalr.Hub
}

func (m *SignalR) UpdateContentInfo(data payload.InfoStat) {
}

func (m *SignalR) Broadcast(eventType payload.EventType, data interface{}) {
}

func (m *SignalR) SizeUpdate(id string, size string) {
}

func (m *SignalR) ProgressUpdate(data payload.ContentProgressUpdate) {
}

func (m *SignalR) StateUpdate(id string, state payload.ContentState) {
}

func (m *SignalR) AddContent(data payload.InfoStat) {
}

func (m *SignalR) DeleteContent(id string) {
}

func (m *SignalR) Notify(models.Notification) {}
