package metadata

import (
	"strconv"
	"strings"

	"github.com/Fesaa/Media-Provider/utils"
)

const (
	Version SemanticVersion = "0.3.3"
)

type SemanticVersion string

func (v SemanticVersion) String() string {
	return string(v)
}

func (v SemanticVersion) Older(v2 SemanticVersion) bool {
	return compareSemanticVersions(string(v), string(v2)) < 0
}

func (v SemanticVersion) Newer(v2 SemanticVersion) bool {
	return compareSemanticVersions(string(v), string(v2)) > 0
}

func (v SemanticVersion) Equal(v2 SemanticVersion) bool {
	return compareSemanticVersions(string(v), string(v2)) == 0
}

func (v SemanticVersion) EqualS(v2 string) bool {
	return v.Equal(SemanticVersion(v2))
}

func compareSemanticVersions(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	if len(parts1) != 3 || len(parts2) != 3 {
		panic("invalid version string")
	}

	major1 := utils.MustReturn(strconv.Atoi(parts1[0]))
	minor1 := utils.MustReturn(strconv.Atoi(parts1[1]))
	patch1 := utils.MustReturn(strconv.Atoi(parts1[2]))

	major2 := utils.MustReturn(strconv.Atoi(parts2[0]))
	minor2 := utils.MustReturn(strconv.Atoi(parts2[1]))
	patch2 := utils.MustReturn(strconv.Atoi(parts2[2]))

	if major1 < major2 {
		return -1
	} else if major1 > major2 {
		return 1
	}

	if minor1 < minor2 {
		return -1
	} else if minor1 > minor2 {
		return 1
	}

	if patch1 < patch2 {
		return -1
	} else if patch1 > patch2 {
		return 1
	}

	return 0
}
