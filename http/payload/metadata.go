package payload

import (
	"github.com/Fesaa/Media-Provider/config"
	"time"
)

type Metadata struct {
	Version               config.SemanticVersion `json:"version"`
	FirstInstalledVersion string                 `json:"firstInstalledVersion"`
	InstallDate           time.Time              `json:"installDate"`
}
