package web

import "embed"

//go:embed index.html css js all:assets
var FS embed.FS
