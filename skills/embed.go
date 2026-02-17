package skills

import (
	"embed"
	"io/fs"
)

//go:embed **/SKILL.md
var embeddedFS embed.FS

// DefaultFS returns the embedded default skills filesystem.
func DefaultFS() (fs.FS, error) {
	return embeddedFS, nil
}
