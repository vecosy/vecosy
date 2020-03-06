package security

import "errors"

// ErrNoMetadataFound will be return in case of no metadata found on the GRPC request
var ErrNoMetadataFound = errors.New("no metadata found")

// ErrAuthFailed will be return in case of some authentication issue
var ErrAuthFailed = errors.New("authentication failed")
