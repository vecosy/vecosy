package configrepo

import "fmt"

var FileNotFoundError = fmt.Errorf("file not found")
var ApplicationNotFoundError = fmt.Errorf("application not found")
