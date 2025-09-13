package metadata

import "fmt"

var (
	CommitHash     = "N/A"
	BuildTimestamp = "N/A"
)

var (
	Identifier = fmt.Sprintf("media-provider@%s+%s", Version, CommitHash)
)
