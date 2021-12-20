package resources

import (
	"embed"
)

// This file only exists because go embed doesn't allow relative paths. So we build the embedded files into this package
// and then use the filesystem it exports

//go:embed *.html
var Res embed.FS
