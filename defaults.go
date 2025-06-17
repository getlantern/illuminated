package illuminated

import (
	"os"
)

var (
	DefaultFilePermissions   = os.FileMode(0o750)
	DefaultFileNameOverrides = "overrides.yml"
	DefaultDirNameStaging    = "staging"
	DefaultDirNameOutput     = "output"
)
