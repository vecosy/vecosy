package validation

import (
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func ValidateApplicationVersion(app *configrepo.ApplicationVersion) error {
	if _, err := ParseVersion(app.AppVersion); err != nil {
		return err
	}
	if app.AppName == "" {
		return InvalidApplicationName
	}
	return nil
}
