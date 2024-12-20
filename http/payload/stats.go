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
	Provider    models.Provider `json:"provider"`
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Size        string          `json:"size"`
	Downloading bool            `json:"downloading"`
	Progress    int64           `json:"progress"`
	Estimated   *int64          `json:"estimated,omitempty"`
	SpeedType   SpeedType       `json:"speed_type"`
	Speed       SpeedData       `json:"speed"`
	DownloadDir string          `json:"download_dir"`
}

type SpeedType int

const (
	BYTES SpeedType = iota
	VOLUMES
	IMAGES
)

type SpeedData struct {
	T     int64 `json:"time"`
	Speed int64 `json:"speed"`
}
