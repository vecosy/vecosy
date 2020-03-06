package validation

import (
	"github.com/hashicorp/go-version"
)

// ParseVersion convert a string version the the version.Version struct
func ParseVersion(strVersion string) (*version.Version, error) {
	ver, err := version.NewVersion(strVersion)
	if err != nil {
		return nil, ErrInvalidVersion
	}
	return ver, nil
}
