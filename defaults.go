package illuminated

import (
	"os"
)

var (
	DefaultDirProject        = "docs"
	DefaultDirNameStaging    = "staging"
	DefaultDirNameOutput     = "output"
	DefaultFileNameOverrides = "overrides.yml"
	DefaultFilePermissions   = os.FileMode(0o750)
)
