package web

import "embed"

//go:embed index.html manifest.webmanifest css js assets
var FS embed.FS
