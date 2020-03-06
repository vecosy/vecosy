package validation

import "errors"

// ErrInvalidVersion returned if the application version is not valid
var ErrInvalidVersion = errors.New("invalid version")

// ErrInvalidApplicationName returned if the application has an invalid name
var ErrInvalidApplicationName = errors.New("invalid application name")
