package payload

import "github.com/Fesaa/Media-Provider/config"

type QueueStat struct {
	Provider config.Provider `json:"provider"`
	Id       string          `json:"id"`
	Name     string          `json:"name,omitempty"`
	BaseDir  string
}

type InfoStat struct {
	Provider    config.Provider `json:"provider"`
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Size        string          `json:"size"`
	Progress    int64           `json:"progress"`
	SpeedType   SpeedType       `json:"speed_type"`
	Speed       SpeedData       `json:"speed"`
	DownloadDir string          `json:"download_dir"`
}

type SpeedType string

const (
	BYTES   SpeedType = "bytes"
	VOLUMES SpeedType = "volumes"
	IMAGES  SpeedType = "images"
)

type SpeedData struct {
	T     int64 `json:"time"`
	Speed int64 `json:"speed"`
}
