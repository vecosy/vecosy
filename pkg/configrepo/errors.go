package configrepo

import "fmt"

// ErrFileNotFound returned if the file has not be found on the repo
var ErrFileNotFound = fmt.Errorf("file not found")

// ErrApplicationNotFound returned if the requested application has not been found on the repo
var ErrApplicationNotFound = fmt.Errorf("application not found")
