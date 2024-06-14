package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
)

type TorrentInfo struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Date        string `json:"Date"`
	Size        string `json:"Size"`
	Seeders     string `json:"Seeders"`
	Leechers    string `json:"Leechers"`
	Downloads   string `json:"Downloads"`
	Link        string `json:"Link"`
	InfoHash    string `json:"InfoHash"`
	ImageUrl    string `json:"ImageUrl"`
	RefUrl      string `json:"RefUrl"`
}

type SearchRequest struct {
	Provider config.Provider `json:"provider"`
	Query    string          `json:"query"`
	Category string          `json:"category,omitempty"`
	SortBy   string          `json:"sort_by,omitempty"`
	Filter   string          `json:"filter,omitempty"`
}

type DownloadRequest struct {
	Provider config.Provider `json:"provider"`
	Hash     string          `json:"info"`
	BaseDir  string          `json:"base_dir"`
}

func (d DownloadRequest) DebugString() string {
	return fmt.Sprintf("{Hash: %s, BaseDir: %s, Url: %t}", d.Hash, d.BaseDir)
}

type StopRequest struct {
	Provider    config.Provider `json:"provider"`
	Id          string          `json:"id"`
	DeleteFiles bool            `json:"delete_files"`
}

type provider interface {
	Search(request SearchRequest) ([]TorrentInfo, error)
	Download(request DownloadRequest) error
	Stop(StopRequest) error
}
