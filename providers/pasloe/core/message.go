package core

import (
	"encoding/json"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"strconv"
)

func (c *Core[C, S]) Message(msg payload.Message) (payload.Message, error) {
	var jsonBytes []byte
	var err error
	switch msg.MessageType {
	case payload.MessageListContent:
		jsonBytes, err = json.Marshal(c.ContentList())
	case payload.SetToDownload:
		err = c.SetUserFiltered(msg.Data)
	case payload.StartDownload:
		err = c.MarkReady()
	default:
		return payload.Message{}, services.ErrUnknownMessageType
	}

	if err != nil {
		return payload.Message{}, err
	}

	return payload.Message{
		Provider:    c.Req.Provider,
		ContentId:   c.id,
		MessageType: msg.MessageType,
		Data:        jsonBytes,
	}, nil
}

func (c *Core[C, S]) MarkReady() error {
	if c.contentState != payload.ContentStateWaiting {
		return services.ErrWrongState
	}

	c.SetState(payload.ContentStateReady)
	return c.Client.MoveToDownloadQueue(c.id)
}

func (c *Core[C, S]) SetUserFiltered(msg json.RawMessage) error {
	if c.contentState != payload.ContentStateWaiting &&
		c.contentState != payload.ContentStateReady {
		return services.ErrWrongState
	}

	var filter []string
	err := json.Unmarshal(msg, &filter)
	if err != nil {
		return err
	}
	c.ToDownloadUserSelected = filter
	c.SignalR.SizeUpdate(c.Id(), strconv.Itoa(c.Size())+" Chapters")
	return nil
}
