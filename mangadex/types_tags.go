package mangadex

type TagResponse MangaDexResponse[TagData]

type TagData struct {
	Id         string        `json:"id"`
	Type       string        `json:"type"`
	Attributes TagAttributes `json:"attributes"`
}

type TagAttributes struct {
	Name          map[string]string `json:"name"`
	Description   map[string]string `json:"description"`
	Group         string            `json:"group"`
	Version       int               `json:"version"`
	Relationships []Relationship    `json:"relationships"`
}
