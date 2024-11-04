package webtoon

import "github.com/Fesaa/Media-Provider/http/payload"

type WebToon interface {
	Title() string
	Id() string
	GetBaseDir() string
	Downloading() bool
	Cancel()
	WaitForInfoAndDownload()
	GetInfo() payload.InfoStat
	GetDownloadDir() string
	GetPrevChapters() []string
}

type SearchOptions struct {
	Query string
}

type Response struct {
	Result  SearchResult `json:"result"`
	Success bool         `json:"success"`
}

type SearchResult struct {
	Query        string       `json:"query"`
	Start        int          `json:"start"`
	Display      int          `json:"display"`
	Total        int          `json:"total"`
	SearchedList []SearchData `json:"searchedList"`
}

type SearchData struct {
	Id              int      `json:"titleNo"`
	Name            string   `json:"title"`
	ReadCount       int      `json:"readCount"`
	ThumbnailMobile string   `json:"thumbnailMobile"`
	AuthorNameList  []string `json:"authorNameList"`
	Genre           string   `json:"representGenre"`
	Rating          bool     `json:"titleUnsuitableForChildren"`
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
