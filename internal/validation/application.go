package validation

import (
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

// ValidateApplicationVersion validate the application name and version
func ValidateApplicationVersion(app *configrepo.ApplicationVersion) error {
	if _, err := ParseVersion(app.AppVersion); err != nil {
		return err
	}
	if app.AppName == "" {
		return ErrInvalidApplicationName
	}
	return nil
}
