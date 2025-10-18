package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
)

type QueueStat struct {
	Provider models.Provider `json:"provider"`
	Id       string          `json:"id"`
	Name     string          `json:"name,omitempty"`
	BaseDir  string
}

type InfoStat struct {
	Provider     models.Provider `json:"provider"`
	Id           string          `json:"id"`
	ContentState ContentState    `json:"contentState"`
	Name         string          `json:"name"`
	RefUrl       string          `json:"ref_url"`
	Size         string          `json:"size"`
	Downloading  bool            `json:"downloading"`
	Progress     int64           `json:"progress"`
	Estimated    int64           `json:"estimated,omitempty"`
	SpeedType    SpeedType       `json:"speed_type"`
	Speed        int64           `json:"speed"`
	DownloadDir  string          `json:"download_dir"`
}

type ContentState int

const (
	// ContentStateQueued indicates the content cannot start retrieving information yet
	ContentStateQueued ContentState = iota
	// ContentStateLoading indicates the content is still retrieving the information needed to start downloading
	ContentStateLoading
	// ContentStateWaiting indicates the content was prevented from downloaded imitatively, and has loaded all information
	// to start downloading
	ContentStateWaiting
	// ContentStateReady indicates the content has been marked for download, but cannot start downloading yet
	ContentStateReady
	// ContentStateDownloading indicates the content is currently being downloaded
	ContentStateDownloading
	// ContentStateCleanup indicates the content is being zipped
	ContentStateCleanup
)

type SpeedType int

const (
	BYTES SpeedType = iota
	VOLUMES
	IMAGES
)
