package webtoon

import "github.com/Fesaa/Media-Provider/payload"

type Config interface {
	GetRootDir() string
	GetMaxConcurrentMangadexImages() int
}

type Client interface {
	Download(payload.DownloadRequest) (WebToon, error)
	RemoveDownload(payload.StopRequest) error
	GetBaseDir() string
	GetCurrentWebToon() WebToon
	GetQueuedWebToons() []payload.QueueStat
	GetConfig() Config
}

type WebToon interface {
	Title() string
	Id() string
	GetBaseDir() string
	Cancel()
	WaitForInfoAndDownload()
	GetInfo() payload.InfoStat
	GetDownloadDir() string
	GetPrevChapters() []string
}

type SearchOptions struct {
	Query string
}

type SearchData struct {
	Id       string
	Name     string
	Author   string
	Genre    string
	ImageUrl string
	Url      string
}

type Series struct {
	Id          string
	Name        string
	Author      string
	Description string
	Genre       string
	Completed   bool
	Chapters    []Chapter
}

type Chapter struct {
	Url      string
	ImageUrl string
	Title    string
	Number   string
	Date     string
}
