package payload

import (
	"github.com/Fesaa/Media-Provider/metadata"
	"time"
)

type Metadata struct {
	Version               metadata.SemanticVersion `json:"version"`
	FirstInstalledVersion string                   `json:"firstInstalledVersion"`
	InstallDate           time.Time                `json:"installDate"`
}
