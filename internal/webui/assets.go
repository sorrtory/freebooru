// Package webui embeds and serves the FreeBooru Vue production bundle.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// Assets returns the embedded production bundle rooted at its public paths.
func Assets() (fs.FS, error) {
	return fs.Sub(embedded, "dist")
}
