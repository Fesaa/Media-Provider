package payload

import (
	"time"
)

type Metadata struct {
	Version               string    `json:"version"`
	FirstInstalledVersion string    `json:"firstInstalledVersion"`
	InstallDate           time.Time `json:"installDate"`
}
