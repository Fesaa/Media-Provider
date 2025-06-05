package core

import (
	"encoding/json"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"strconv"
)

func (d *DownloadBase[T]) Message(msg payload.Message) (payload.Message, error) {
	var jsonBytes []byte
	var err error
	switch msg.MessageType {
	case payload.MessageListContent:
		jsonBytes, err = json.Marshal(d.infoProvider.ContentList())
	case payload.SetToDownload:
		err = d.SetUserFiltered(msg.Data)
	case payload.StartDownload:
		err = d.MarkReady()
	default:
		return payload.Message{}, services.ErrUnknownMessageType
	}

	if err != nil {
		return payload.Message{}, err
	}

	return payload.Message{
		Provider:    d.Req.Provider,
		ContentId:   d.id,
		MessageType: msg.MessageType,
		Data:        jsonBytes,
	}, nil
}

func (d *DownloadBase[T]) MarkReady() error {
	if d.contentState != payload.ContentStateWaiting {
		return services.ErrWrongState
	}

	if d.Client.CanStart(d.Req.Provider) {
		go d.StartDownload()
		return nil
	}

	d.SetState(payload.ContentStateReady)
	return nil
}

func (d *DownloadBase[T]) SetUserFiltered(msg json.RawMessage) error {
	if d.contentState != payload.ContentStateWaiting &&
		d.contentState != payload.ContentStateReady {
		return services.ErrWrongState
	}

	var filter []string
	err := json.Unmarshal(msg, &filter)
	if err != nil {
		return err
	}
	d.ToDownloadUserSelected = filter
	d.SignalR.SizeUpdate(d.Id(), strconv.Itoa(d.Size())+" Chapters")
	return nil
}
