package embedded

import (
	"embed"
	"io/fs"
)

//go:embed build
var buildFS embed.FS

func FS() fs.FS {
	sub, err := fs.Sub(buildFS, "build")
	if err != nil {
		panic(err)
	}
	return sub
}
