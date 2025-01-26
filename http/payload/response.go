package payload

type StatsResponse struct {
	Running []InfoStat  `json:"running"`
	Queued  []QueueStat `json:"queued"`
}

type ListDirResponse []DirEntry

type DirEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
}

type DownloadMetadata struct {
	Definitions []DownloadMetadataDefinition `json:"definitions"`
}

type DownloadMetadataDefinition struct {
	Title    string                   `json:"title"`
	Key      string                   `json:"key"`
	FormType DownloadMetadataFormType `json:"formType"`
	Required bool                     `json:"required"`
	Options  []MetadataOption         `json:"options"`
}

type MetadataOption struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DownloadMetadataFormType int

const (
	SWITCH DownloadMetadataFormType = iota
	DROPDOWN
	MULTI
)
