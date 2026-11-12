package assets

import "embed"

// Files contains the application templates and static assets.
//
//go:embed templates/*.html static
var Files embed.FS
