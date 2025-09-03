package payload

import (
	"time"

	"github.com/Fesaa/Media-Provider/metadata"
)

type Metadata struct {
	Version               metadata.SemanticVersion `json:"version"`
	FirstInstalledVersion string                   `json:"firstInstalledVersion"`
	InstallDate           time.Time                `json:"installDate"`
}
