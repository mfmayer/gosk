package utils

import (
	"io/fs"
	"os"
)

func FilesExist(fsys fs.FS, filePaths ...string) bool {
	for _, path := range filePaths {
		if _, err := fs.Stat(fsys, path); err != nil {
			if os.IsNotExist(err) {
				return false
			}
		}
	}
	return true
}
