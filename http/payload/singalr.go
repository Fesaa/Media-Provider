package payload

type EventType string

const (
	EventTypeContentSizeUpdate     EventType = "ContentSizeUpdate"
	EventTypeContentProgressUpdate EventType = "ContentProgressUpdate"
	EventTypeAddContent            EventType = "AddContent"
	EventTypeDeleteContent         EventType = "DeleteContent"
	EventTypeContentStateUpdate    EventType = "ContentStateUpdate"
)

type ContentSizeUpdate struct {
	ContentId string `json:"contentId"`
	Size      string `json:"size"`
}

type ContentProgressUpdate struct {
	ContentId string    `json:"contentId"`
	Progress  int64     `json:"progress"`
	Estimated *int64    `json:"estimated,omitempty"`
	SpeedType SpeedType `json:"speed_type"`
	Speed     int64     `json:"speed"`
}

type DeleteContent struct {
	ContentId string `json:"contentId"`
}

type ContentStateUpdate struct {
	ContentId    string       `json:"contentId"`
	ContentState ContentState `json:"contentState"`
}
