package illuminated

import (
	"os"
)

var (
	DefaultFilePermissions = os.FileMode(0o750)
	DefaultDirNameStaging  = "staging"
	DefaultDirNameOutput   = "output"
)
