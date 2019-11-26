package validation

import (
	"github.com/hashicorp/go-version"
)

func ParseVersion(strVersion string) (*version.Version, error) {
	ver, err := version.NewVersion(strVersion)
	if err != nil {
		return nil, InvalidVersion
	}
	return ver, nil
}
