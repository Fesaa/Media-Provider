package payload

import (
	"encoding/json"
	"github.com/Fesaa/Media-Provider/db/models"
)

type Message struct {
	Provider    models.Provider `json:"provider"`
	ContentId   string          `json:"contentId"`
	MessageType MessageType     `json:"type"`
	Data        json.RawMessage `json:"data,omitempty"`
}

type MessageType int

const (
	MessageListContent MessageType = iota
	SetToDownload
	StartDownload
)

type ListContentData struct {
	SubContentId string            `json:"subContentId,omitempty"`
	Label        string            `json:"label"`
	Selected     bool              `json:"selected"`
	Children     []ListContentData `json:"children,omitempty"`
}
