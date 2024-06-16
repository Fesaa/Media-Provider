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
