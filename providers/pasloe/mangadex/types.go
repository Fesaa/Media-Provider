package mangadex

type SearchOptions struct {
	Query                  string
	IncludedTags           []string
	ExcludedTags           []string
	Status                 []string
	ContentRating          []string
	PublicationDemographic []string
	SkipNotFoundTags       bool
}

type Response[T any] struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     T      `json:"data"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Total    int    `json:"total"`
}

type Relationship struct {
	Id         string         `json:"id"`
	Type       string         `json:"type"`
	Related    string         `json:"related,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}
