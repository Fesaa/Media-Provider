package payload

import (
	"encoding/json"
	"github.com/Fesaa/Media-Provider/db/models"
)

type Message struct {
	Provider    models.Provider `json:"provider"`
	ContentId   string          `json:"contentId"`
	MessageType MessageType     `json:"type"`
	Data        json.RawMessage `json:"data"`
}

type MessageType int
