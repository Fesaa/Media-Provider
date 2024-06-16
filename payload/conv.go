package payload

func (r DownloadRequest) ToQueueStat() QueueStat {
	return QueueStat{
		Provider: r.Provider,
		Id:       r.Id,
		Name:     "",
		BaseDir:  r.BaseDir,
	}
}

func (q QueueStat) ToDownloadRequest() DownloadRequest {
	return DownloadRequest{
		Provider: q.Provider,
		Id:       q.Id,
		BaseDir:  q.BaseDir,
	}
}
